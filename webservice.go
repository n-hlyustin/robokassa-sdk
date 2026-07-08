package robokassa

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []xmlNode  `xml:",any"`
}

func nodeToMap(node xmlNode) interface{} {
	if len(node.Attrs) == 0 && len(node.Nodes) == 0 {
		return node.Content
	}
	result := map[string]interface{}{}
	for _, attr := range node.Attrs {
		result[attr.Name.Local] = attr.Value
	}
	if strings.TrimSpace(node.Content) != "" {
		result["Content"] = node.Content
	}
	for _, child := range node.Nodes {
		childValue := nodeToMap(child)
		if existing, ok := result[child.XMLName.Local]; ok {
			switch values := existing.(type) {
			case []interface{}:
				result[child.XMLName.Local] = append(values, childValue)
			default:
				result[child.XMLName.Local] = []interface{}{values, childValue}
			}
			continue
		}
		result[child.XMLName.Local] = childValue
	}
	return result
}

type WebService struct {
	transport     *transport
	signer        signer
	merchantLogin string
	password2     string
	hashType      string
	webServiceURL string
	currenciesURL string
}

func (s *WebService) GetPaymentMethods(ctx context.Context, lang string) (*PaymentMethodsResponse, error) {
	if lang == "" {
		lang = "en"
	}
	query := url.Values{}
	query.Set("MerchantLogin", s.merchantLogin)
	query.Set("Language", lang)
	endpoint := s.webServiceURL + "/GetPaymentMethods?" + query.Encode()

	resp, err := s.transport.get(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, &SDKError{Op: "webservice.get_payment_methods", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	var root xmlNode
	if err := xml.Unmarshal(resp.Body, &root); err != nil {
		return nil, &SDKError{Op: "webservice.get_payment_methods", Message: "failed to parse XML", Err: err}
	}
	data, _ := nodeToMap(root).(map[string]interface{})
	return &PaymentMethodsResponse{
		RawXML:  string(resp.Body),
		Result:  extractXMLResult(data),
		Groups:  extractPaymentMethodGroups(data),
		RawData: data,
	}, nil
}

func (s *WebService) GetCurrencies(ctx context.Context, lang string) (*CurrenciesResponse, error) {
	if lang == "" {
		lang = "en"
	}
	query := url.Values{}
	query.Set("MerchantLogin", s.merchantLogin)
	query.Set("Language", lang)
	endpoint := s.currenciesURL + "/GetCurrencies?" + query.Encode()

	resp, err := s.transport.get(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, &SDKError{Op: "webservice.get_currencies", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	var root xmlNode
	if err := xml.Unmarshal(resp.Body, &root); err != nil {
		return nil, &SDKError{Op: "webservice.get_currencies", Message: "failed to parse XML", Err: err}
	}
	data, _ := nodeToMap(root).(map[string]interface{})
	return &CurrenciesResponse{
		RawXML:  string(resp.Body),
		Result:  extractXMLResult(data),
		Groups:  extractCurrencyGroups(data),
		RawData: data,
	}, nil
}

func (s *WebService) OpState(ctx context.Context, invoiceID int64) (*OpStateResponse, error) {
	signatureValue, err := s.signer.SignOpState(s.merchantLogin, fmt.Sprintf("%d", invoiceID), s.password2, s.hashType)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("MerchantLogin", s.merchantLogin)
	query.Set("InvoiceID", fmt.Sprintf("%d", invoiceID))
	query.Set("Signature", signatureValue)
	endpoint := s.webServiceURL + "/OpStateExt?" + query.Encode()

	resp, err := s.transport.get(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if resp.Status != 200 {
		return nil, &SDKError{Op: "webservice.op_state", StatusCode: resp.Status, Message: "unexpected HTTP status"}
	}

	var root xmlNode
	if err := xml.Unmarshal(resp.Body, &root); err != nil {
		return nil, &SDKError{Op: "webservice.op_state", Message: "failed to parse XML", Err: err}
	}
	data, _ := nodeToMap(root).(map[string]interface{})
	return &OpStateResponse{
		RawXML:  string(resp.Body),
		Result:  extractXMLResult(data),
		State:   extractInvoiceState(data),
		RawData: data,
	}, nil
}

func extractXMLResult(data map[string]interface{}) XMLResult {
	result := XMLResult{}
	for _, key := range []string{"Result", "result"} {
		if value, ok := data[key].(map[string]interface{}); ok {
			result.Code = fmt.Sprint(firstMapValue(value, "Code", "code"))
			result.Description = fmt.Sprint(firstMapValue(value, "Description", "description"))
			return result
		}
	}
	return result
}

func extractPaymentMethodGroups(data map[string]interface{}) []PaymentMethodGroup {
	if methods := extractPaymentMethods(firstMapValue(data, "Methods", "methods")); len(methods) > 0 {
		return []PaymentMethodGroup{{Items: methods}}
	}

	values := flattenMapsByKey(data, "Groups", "Group", "Methods", "Method")
	out := make([]PaymentMethodGroup, 0)
	for _, value := range values {
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		group := PaymentMethodGroup{
			Code:        stringify(firstMapValue(entry, "Code", "code")),
			Description: stringify(firstMapValue(entry, "Description", "description", "Name", "name")),
		}
		for _, item := range flattenMapsByKey(entry, "Items", "Item", "Methods", "Method") {
			methodMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			group.Items = append(group.Items, PaymentMethod{
				Code:        stringify(firstMapValue(methodMap, "Code", "code")),
				Description: stringify(firstMapValue(methodMap, "Description", "description", "Name", "name")),
			})
		}
		if group.Code != "" || group.Description != "" || len(group.Items) > 0 {
			out = append(out, group)
		}
	}
	return out
}

func extractPaymentMethods(value interface{}) []PaymentMethod {
	methodsMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	rawMethods, ok := methodsMap["Method"]
	if !ok {
		rawMethods, ok = methodsMap["method"]
	}
	if !ok {
		return nil
	}

	values := []interface{}{rawMethods}
	if items, ok := rawMethods.([]interface{}); ok {
		values = items
	}

	out := make([]PaymentMethod, 0, len(values))
	for _, value := range values {
		methodMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		method := PaymentMethod{
			Code:        stringify(firstMapValue(methodMap, "Code", "code")),
			Description: stringify(firstMapValue(methodMap, "Description", "description", "Name", "name")),
		}
		if method.Code != "" || method.Description != "" {
			out = append(out, method)
		}
	}
	return out
}

func extractCurrencyGroups(data map[string]interface{}) []CurrencyGroup {
	if groups := extractCurrencyGroupsFromValue(firstMapValue(data, "Groups", "groups")); len(groups) > 0 {
		return groups
	}
	if currencies := extractCurrencies(firstMapValue(data, "Currencies", "currencies")); len(currencies) > 0 {
		return []CurrencyGroup{{Items: currencies}}
	}
	return nil
}

func extractCurrencyGroupsFromValue(value interface{}) []CurrencyGroup {
	groupsMap, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}
	rawGroups, ok := groupsMap["Group"]
	if !ok {
		rawGroups, ok = groupsMap["group"]
	}
	if !ok {
		return nil
	}

	values := []interface{}{rawGroups}
	if items, ok := rawGroups.([]interface{}); ok {
		values = items
	}

	out := make([]CurrencyGroup, 0, len(values))
	for _, value := range values {
		groupMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		group := CurrencyGroup{
			Code:        stringify(firstMapValue(groupMap, "Code", "code")),
			Description: stringify(firstMapValue(groupMap, "Description", "description", "Name", "name")),
		}
		group.Items = extractCurrencies(firstMapValue(groupMap, "Items", "items"))
		if len(group.Items) == 0 {
			group.Items = extractCurrencies(firstMapValue(groupMap, "Currencies", "currencies"))
		}
		if group.Code != "" || group.Description != "" || len(group.Items) > 0 {
			out = append(out, group)
		}
	}
	return out
}

func extractCurrencies(value interface{}) []Currency {
	values := []interface{}{value}
	if items, ok := value.([]interface{}); ok {
		values = items
	}

	out := make([]Currency, 0, len(values))
	for _, value := range values {
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		rawCurrencies, ok := entry["Currency"]
		if !ok {
			rawCurrencies, ok = entry["currency"]
		}
		if ok {
			out = append(out, extractCurrencies(rawCurrencies)...)
			continue
		}

		currency := Currency{
			Label: stringify(firstMapValue(entry, "Label", "label", "Code", "code")),
			Name:  stringify(firstMapValue(entry, "Name", "name", "Description", "description")),
		}
		if currency.Label != "" || currency.Name != "" {
			out = append(out, currency)
		}
	}
	return out
}

func extractInvoiceState(data map[string]interface{}) *InvoiceState {
	candidates := flattenMapsByKey(data, "State", "OperationState", "Info")
	for _, value := range candidates {
		entry, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		state := &InvoiceState{
			InvoiceID:     stringify(firstMapValue(entry, "InvoiceID", "InvoiceId", "InvId", "id")),
			StateCode:     stringify(firstMapValue(entry, "Code", "StateCode", "code")),
			StateName:     stringify(firstMapValue(entry, "State", "Description", "state")),
			OutSum:        stringify(firstMapValue(entry, "OutSum", "outsum")),
			IncSum:        stringify(firstMapValue(entry, "IncSum", "incsum")),
			PaymentMethod: stringify(firstMapValue(entry, "PaymentMethod", "PaymentMethodCode", "PaymentMethodName")),
			OpKey:         stringify(firstMapValue(entry, "OpKey", "OpKeyInv")),
		}
		if *state != (InvoiceState{}) {
			return state
		}
	}
	return nil
}

func flattenMapsByKey(data map[string]interface{}, keys ...string) []interface{} {
	out := make([]interface{}, 0)
	var walk func(interface{})
	walk = func(value interface{}) {
		switch typed := value.(type) {
		case map[string]interface{}:
			for key, child := range typed {
				for _, target := range keys {
					if key == target {
						switch items := child.(type) {
						case []interface{}:
							out = append(out, items...)
						default:
							out = append(out, items)
						}
					}
				}
				walk(child)
			}
		case []interface{}:
			for _, item := range typed {
				walk(item)
			}
		}
	}
	walk(data)
	return out
}

func firstMapValue(m map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := m[key]; ok {
			return value
		}
	}
	return ""
}

func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
