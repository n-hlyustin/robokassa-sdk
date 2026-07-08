package robokassa

import (
	"encoding/xml"
	"testing"
)

func TestExtractPaymentMethodsFromXMLAttributes(t *testing.T) {
	raw := []byte(`<?xml version="1.0" encoding="utf-8"?>
<PaymentMethodsList xmlns="http://merchant.roboxchange.com/WebService/">
	<Result><Code>0</Code></Result>
	<Methods>
		<Method Code="BankCard" Description="Банковской картой" />
	</Methods>
</PaymentMethodsList>`)

	var root xmlNode
	if err := xml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("unexpected XML error: %v", err)
	}

	data, ok := nodeToMap(root).(map[string]interface{})
	if !ok {
		t.Fatal("expected XML root to be converted to map")
	}

	groups := extractPaymentMethodGroups(data)
	if len(groups) != 1 {
		t.Fatalf("expected one payment method group, got %d: %#v", len(groups), groups)
	}
	if len(groups[0].Items) != 1 {
		t.Fatalf("expected one payment method, got %d: %#v", len(groups[0].Items), groups[0].Items)
	}
	if groups[0].Items[0].Code != "BankCard" {
		t.Fatalf("unexpected payment method code: %q", groups[0].Items[0].Code)
	}
	if groups[0].Items[0].Description != "Банковской картой" {
		t.Fatalf("unexpected payment method description: %q", groups[0].Items[0].Description)
	}
}

func TestExtractCurrenciesFromXMLAttributes(t *testing.T) {
	raw := []byte(`<?xml version="1.0" encoding="utf-8"?>
<CurrenciesList xmlns="http://merchant.roboxchange.com/WebService/">
	<Result><Code>0</Code></Result>
	<Groups>
		<Group Code="BankCard" Description="Банковские карты">
			<Items>
				<Currency Label="BANKOCEAN2R" Name="Банковской картой" />
			</Items>
		</Group>
	</Groups>
</CurrenciesList>`)

	var root xmlNode
	if err := xml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("unexpected XML error: %v", err)
	}
	data, ok := nodeToMap(root).(map[string]interface{})
	if !ok {
		t.Fatal("expected XML root to be converted to map")
	}

	groups := extractCurrencyGroups(data)
	if len(groups) != 1 {
		t.Fatalf("expected one currency group, got %d: %#v", len(groups), groups)
	}
	if groups[0].Code != "BankCard" {
		t.Fatalf("unexpected group code: %q", groups[0].Code)
	}
	if len(groups[0].Items) != 1 {
		t.Fatalf("expected one currency, got %d: %#v", len(groups[0].Items), groups[0].Items)
	}
	if groups[0].Items[0].Label != "BANKOCEAN2R" {
		t.Fatalf("unexpected currency label: %q", groups[0].Items[0].Label)
	}
	if groups[0].Items[0].Name != "Банковской картой" {
		t.Fatalf("unexpected currency name: %q", groups[0].Items[0].Name)
	}
}

func TestExtractInvoiceStateFromXML(t *testing.T) {
	raw := []byte(`<?xml version="1.0" encoding="utf-8"?>
<OperationStateResponse xmlns="http://merchant.roboxchange.com/WebService/">
	<Result><Code>0</Code></Result>
	<State>
		<InvoiceID>1001</InvoiceID>
		<Code>100</Code>
		<State>Оплачено</State>
		<OutSum>99.90</OutSum>
		<IncSum>99.90</IncSum>
		<PaymentMethod>BankCard</PaymentMethod>
		<OpKey>abc</OpKey>
	</State>
</OperationStateResponse>`)

	var root xmlNode
	if err := xml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("unexpected XML error: %v", err)
	}
	data, ok := nodeToMap(root).(map[string]interface{})
	if !ok {
		t.Fatal("expected XML root to be converted to map")
	}

	state := extractInvoiceState(data)
	if state == nil {
		t.Fatal("expected invoice state")
	}
	if state.InvoiceID != "1001" {
		t.Fatalf("unexpected invoice id: %q", state.InvoiceID)
	}
	if state.StateCode != "100" {
		t.Fatalf("unexpected state code: %q", state.StateCode)
	}
	if state.PaymentMethod != "BankCard" {
		t.Fatalf("unexpected payment method: %q", state.PaymentMethod)
	}
}
