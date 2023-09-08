package filtersupport_test

/*
import (
	"fmt"
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/filtersupport"
	"github.com/stretchr/testify/assert"
)

func TestParseFilter(t *testing.T) {
	examples := [][2]string{
		{"title pr"},
		{"name pr and userName pr and title pr"},
		{"name.familyName co \"O'Malley\""},
		{"(userName eq \"bjensen\")"},
		{"userName   eq  \"bjensen\"", "userName eq \"bjensen\""},
		{"level gt 12"},
		{"level gt 12.3"},
		{"level eq 123.45e-5"},
		{"emails.type eq \"w o(rk)\""},
		{"userName Eq \"bjensen\"", "userName eq \"bjensen\""},
		{"((userName eq A) or (username eq \"B\")) or username eq C", "((userName eq \"A\") or (username eq \"B\")) or username eq \"C\""},
		{"((userName eq A or username eq \"B\") or (username eq C))", "((userName eq \"A\" or username eq \"B\") or (username eq \"C\"))"},
		{"userName sw \"J\""},
		{"urn:ietf:params:scim:schemas:core:2.0:User:userName sw \"J\""},

		{"meta.lastModified gt \"2011-05-13T04:42:34Z\""},
		{"meta.lastModified ge \"2011-05-13T04:42:34Z\""},
		{"meta.lastModified lt \"2011-05-13T04:42:34Z\""},
		{"meta.lastModified le \"2011-05-13T04:42:34Z\""},
		{"title pr and userType eq \"Employee\""},
		{"title pr or userType eq \"Intern\""},
		{"schemas eq \"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User\""},

		{"userType eq \"Employee\" and (emails.type eq \"work\")"},
		{"userType eq \"Employee\" and emails[type eq \"work\" and value co \"@example.com\"]"},
		{"userType eq \"Employee\" and (emails co \"example.com\" or emails.value co \"example.org\")"},
		{"userType ne \"Employee\" and not (emails co \"example.com\" or emails.value co \"example.org\")"},
		{"emails[type eq \"work\" and value co \"@example.com\"] or ims[type eq \"xmpp\" and value co \"@foo.com\"]"},

		{"name pr and not (first eq \"test\") and another ne \"test\""},
		{"NAME PR AND NOT (FIRST EQ \"t[es]t\") AND ANOTHER NE \"test\"", "NAME pr and not (FIRST eq \"t[es]t\") and ANOTHER ne \"test\""},
		{"name pr or userName pr or title pr"},
		{"emails[type eq work and value ew \"h[exa].org\"]", "emails[type eq \"work\" and value ew \"h[exa].org\"]"},
	}
	for _, example := range examples {
		t.Run(example[0], func(t *testing.T) {
			fmt.Println(fmt.Sprintf("Input:\t%s", example[0]))
			ast, err := filtersupport.ParseFilter(example[0])
			assert.NoError(t, err, "Example not parsed: "+example[0])
			element := *ast
			out := element.String()
			fmt.Println(fmt.Sprintf("Parsed:\t%s", out))
			match := example[1]
			if match == "" {
				match = example[0]
			}
			assert.Equal(t, match, out, "Check expected result matches: %s", match)
		})
	}
}

func TestNegParseTests(t *testing.T) {
	ast, err := filtersupport.ParseFilter("username == blah")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Unsupported comparison operator: ==")
	}

	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("((username pr or quota eq 0) and black eq white")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing close ')' bracket")
	}

	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("username pr or quota eq \"none\") and black eq white")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing open '(' bracket")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("username eq \"none\")")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing open '(' bracket")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("username eq \"none\" and")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Incomplete expression")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("username eq \"none\" or abc")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Incomplete expression")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	// The following test is poorly formed. Expression should be emails[type eq work and value ew "hexa.org"]
	ast, err = filtersupport.ParseFilter("emails[type eq work] ew \"hexa.org\"")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing and/or clause")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("emails[type eq work and value ew \"hexa.org\"")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing close ']' bracket")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("emails[type[sub eq val] eq work and value ew \"hexa.org\"")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: A second '[' was detected while looking for a ']' in a value path filter")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	// This checks if a sufFilter expression is invalid
	ast, err = filtersupport.ParseFilter("(username == \"malformed\")")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Unsupported comparison operator: ==")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	// .value is not currently supported
	ast, err = filtersupport.ParseFilter("emails[type eq val].value eq work")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: expecting space after ']' in value path expression")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("emails.type] eq work")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing open '[' bracket")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("emails.type) eq work and a eq b")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Missing open '(' bracket")
	}
	assert.Nil(t, ast, "No filter should be parsed")

	ast, err = filtersupport.ParseFilter("emails[type == work] and a eq b")
	if err != nil {
		fmt.Println(err.Error())
		assert.EqualError(t, err, "invalid IDQL filter: Unsupported comparison operator: ==")
	}
	assert.Nil(t, ast, "No filter should be parsed")
}
*/
