package apimapiop

import (
	"encoding/xml"
	"log"
	"testing"
)

var fragmentXml = `<fragment>
	<set-variable name="my-value" value="hello from policy" />
    <include-fragment fragment-id="test-from-tf" />
    <base />
    <set-header name="MyHeader">
        <value>@((string)context.Variables["my-value"])</value>
    </set-header>
</fragment>`

type SetHeader struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value"`
}
type Fragment struct {
	Header SetHeader `xml:"set-header"`
}

func TestPolicyFragment(t *testing.T) {
	f := Fragment{}
	_ = xml.Unmarshal([]byte(fragmentXml), &f)
	log.Println(f.Header.Value)
	log.Println(f.Header.Name)
}
