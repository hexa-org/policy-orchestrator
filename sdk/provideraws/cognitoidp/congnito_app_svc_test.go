package cognitoidp_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/sdk/core/idp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/app/cognitoidp"
	"github.com/hexa-org/policy-orchestrator/sdk/provideraws/app/cognitoidp/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestNewAppInfoSvc_WithRealCognitoClient - asserts a new AppInfoSvc
// with actual cognito client. Since it uses random aws credentials, we expect an error
// when using the appInfoSvc to make calls.
func TestNewAppInfoSvc_WithRealCognitoClient(t *testing.T) {
	svc, err := cognitoidp.NewAppInfoSvc(testhelper.AwsCredentialsForTest())
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	applications, err := svc.GetApplications()
	assert.ErrorContains(t, err, "StatusCode: 400")
	assert.Nil(t, applications)
}

func TestGetApplications_ListUserPoolsError(t *testing.T) {
	svc, m := newAppInfoSvcWithMock()
	m.ExpectListUserPools(nil, errors.New("user pools error"))
	_, err := svc.GetApplications()
	assert.ErrorContains(t, err, "user pools error")
}

func TestGetApplications_ListResourceServersError(t *testing.T) {
	svc, m := newAppInfoSvcWithMock()
	m.ExpectListUserPools(testhelper.ListUserPoolsOutput(), nil)
	m.ExpectListResourceServers(testhelper.TestUserPoolId, nil, errors.New("resource servers error"))
	_, err := svc.GetApplications()
	assert.ErrorContains(t, err, "resource servers error")
}

func TestGetApplications_Success(t *testing.T) {
	svc, m := newAppInfoSvcWithMock()
	m.ExpectListUserPools(testhelper.ListUserPoolsOutput(), nil)
	m.ExpectListResourceServers(testhelper.TestUserPoolId, testhelper.ListResourceServersOutput(), nil)
	apps, err := svc.GetApplications()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(apps))
	exp := cognitoidp.NewResourceServerAppInfo(testhelper.TestUserPoolId, testhelper.TestResourceServerName, testhelper.TestResourceServerName, testhelper.TestResourceServerIdentifier)
	assert.Equal(t, exp, apps[0].(cognitoidp.ResourceServerAppInfo))
}

func newAppInfoSvcWithMock() (idp.AppInfoSvc, *testhelper.MockCognitoClient) {
	m := testhelper.NewMockCognitoClient()
	svc, _ := cognitoidp.NewAppInfoSvc(nil, cognitoidp.WithCognitoClientOverride(m))
	return svc, m
}
