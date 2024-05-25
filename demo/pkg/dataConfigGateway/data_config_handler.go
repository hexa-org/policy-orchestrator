// This code based on contributions from https://github.com/i2-open/i2goSignals with permission
package dataConfigGateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hexa-org/policy-mapper/api/policyprovider"
	"github.com/hexa-org/policy-mapper/sdk"
	"github.com/hexa-org/policy-orchestrator/demo/pkg/migrationSupport"
)

const EnvIntegrationConfigFile string = "ORCHESTRATOR_CONFIG_FILE"

var ConfigFile = "config.json"

type ConfigData struct {
	ConfigFile   string                      `json:"-"`
	Integrations map[string]*sdk.Integration `json:"integrations"`
	AppData      ApplicationData             `json:"-"`
}

func NewIntegrationConfigData() (*ConfigData, error) {
	config := ConfigData{Integrations: make(map[string]*sdk.Integration)}
	err := config.Load(os.Getenv(EnvIntegrationConfigFile))
	config.AppData = ApplicationData{&config}
	return &config, err
}

func (c *ConfigData) GetApplicationDataGateway() ApplicationsDataGateway {
	return &c.AppData
}

func (c *ConfigData) GetIntegration(alias string) *sdk.Integration {
	integration, exist := c.Integrations[alias]
	if exist {
		return integration
	}
	return nil
}

func (c *ConfigData) GetApplicationInfo(applicationAlias string) (*sdk.Integration, *policyprovider.ApplicationInfo) {
	for _, integration := range c.Integrations {
		app, exist := integration.Apps[applicationAlias]
		if exist {
			return integration, &app
		}
		// Check for match by object id
		for _, app := range integration.Apps {
			if app.ObjectID == applicationAlias {
				return integration, &app
			}
		}
	}
	return nil, nil
}

func (c *ConfigData) checkConfigPath(configPath string) error {

	if configPath == "" {
		configPath = ".hexa/" + ConfigFile
		usr, err := user.Current()
		if err == nil {
			configPath = filepath.Join(usr.HomeDir, configPath)
		}
	}

	dirPath := filepath.Dir(configPath)
	i := len(dirPath)
	if dirPath[i-1:i-1] != "/" {
		dirPath = dirPath + "/"
	}

	pathStat, err := os.Stat(configPath)
	if pathStat != nil && pathStat.IsDir() {
		dirPath = configPath
		configPath = filepath.Join(dirPath, ConfigFile)
	} else {
		_, err = os.Stat(dirPath)
		if os.IsNotExist(err) {

			// path/to/whatever does not exist
			err = os.Mkdir(dirPath, 0770)
			if err != nil {
				return err
			}
		}
	}

	c.ConfigFile = configPath

	return nil
}

func (c *ConfigData) Load(configPath string) error {
	// configFile := filepath.Join(g.Config, ConfigFile)
	c.checkConfigPath(configPath)
	if _, err := os.Stat(c.ConfigFile); os.IsNotExist(err) {
		return nil // No existing configuration
	}

	configBytes, err := os.ReadFile(c.ConfigFile)
	if err != nil {
		fmt.Println("Error reading configuration: " + err.Error())
		return nil
	}
	if len(configBytes) == 0 {
		return nil
	}
	err = json.Unmarshal(configBytes, c)
	if err != nil {
		fmt.Println("Error parsing stored configuration: " + err.Error())
	}
	return err
}

func (c *ConfigData) Save() error {

	out, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(c.ConfigFile, out, 0660)
	if err != nil {
		fmt.Println("Error saving configuration: " + err.Error())
	}
	return err
}

func (c *ConfigData) Create(alias string, providerType string, key []byte) (string, error) {
	mapType := migrationSupport.MapSdkProviderName(providerType)
	integration, err := sdk.OpenIntegration(sdk.WithIntegrationInfo(policyprovider.IntegrationInfo{
		Name: mapType,
		Key:  key,
	}))
	if err != nil {
		return "", err
	}
	if alias == "" {
		alias = generateAliasOfSize(3)
	}

	integration.Alias = alias

	_, err = integration.GetPolicyApplicationPoints(func() string {
		return generateAliasOfSize(4)
	})
	if err != nil {
		return "", err
	}

	c.Integrations[integration.Alias] = integration
	c.Save()
	return integration.Alias, err
}

func (c *ConfigData) Find() []IntegrationRecord {
	resp := make([]IntegrationRecord, 0)
	for _, integration := range c.Integrations {
		resp = append(resp, mapIntegrationRecord(integration))
	}
	return resp
}

func (c *ConfigData) FindById(id string) (IntegrationRecord, error) {
	integration, exist := c.Integrations[id]
	if !exist {
		return IntegrationRecord{}, errors.New("integration does not exist")
	}
	return mapIntegrationRecord(integration), nil
}

func (c *ConfigData) Delete(name string) error {
	_, exists := c.Integrations[name]
	if !exists {
		return errors.New("integration does not exist")
	}
	delete(c.Integrations, name)
	return c.Save()
}

type ApplicationData struct {
	data *ConfigData
}

func (a ApplicationData) Find(refresh bool) ([]ApplicationRecord, error) {
	resp := make([]ApplicationRecord, 0)

	for _, integration := range a.data.Integrations {
		if refresh {
			_, err := integration.GetPolicyApplicationPoints(nil)
			if err != nil {
				return nil, err
			}
		}
		for alias, app := range integration.Apps {
			resp = append(resp, mapApplication(alias, integration, app))
		}
	}

	// Sort by App Name
	sort.Slice(resp, func(i, j int) bool {
		return resp[i].Name < resp[j].Name
	})

	return resp, nil
}

func (a ApplicationData) FindByObjectId(objectId string) (*ApplicationRecord, error) {
	// GetApplicationInfo works on id or objectid
	return a.FindById(objectId)
}

func (a ApplicationData) FindById(id string) (*ApplicationRecord, error) {
	integration, app := a.data.GetApplicationInfo(id)
	if app == nil {
		return nil, errors.New(fmt.Sprintf("application %s not found", id))
	}
	appRec := mapApplication(id, integration, *app)
	return &appRec, nil
}

func (a ApplicationData) DeleteById(id string) error {
	// Comment this used to be important when we were cleaning up database. This doesn't do much, since a refresh auto-populates.
	integration, app := a.data.GetApplicationInfo(id)
	if app == nil {
		return errors.New(fmt.Sprintf("application %s not found", id))
	}
	delete(integration.Apps, id)
	return nil
}

func mapApplication(id string, integ *sdk.Integration, app policyprovider.ApplicationInfo) ApplicationRecord {
	return ApplicationRecord{
		ID:            id,
		IntegrationId: integ.Alias,
		ObjectId:      app.ObjectID,
		Name:          app.Name,
		Description:   app.Description,
		Service:       app.Service,
	}
}

func mapIntegrationRecord(integration *sdk.Integration) IntegrationRecord {
	info := integration.Opts.Info
	return IntegrationRecord{
		ID:       integration.Alias,
		Name:     info.Name,
		Provider: integration.GetType(),
		Key:      info.Key,
	}
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func generateAliasOfSize(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
