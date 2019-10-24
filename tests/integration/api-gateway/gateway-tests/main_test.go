package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/ingressgateway"

	"github.com/avast/retry-go"

	"github.com/stretchr/testify/assert"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/api"
	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/jwt"

	"github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/manifestprocessor"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	manager "github.com/kyma-project/kyma/tests/integration/api-gateway/gateway-tests/pkg/resourcemanager"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const testIDLength = 8
const OauthClientSecretLength = 8
const OauthClientIDLength = 8
const manifestsDirectory = "manifests/"
const testingAppFile = "testing-app.yaml"
const globalCommonResourcesFile = "global-commons.yaml"
const hydraClientFile = "hydra-client.yaml"
const noAccessStrategyApiruleFile = "no_access_strategy.yaml"
const oauthStrategyApiruleFile = "oauth-strategy.yaml"
const jwtAndOauthStrategyApiruleFile = "jwt-oauth-strategy.yaml"
const resourceSeparator = "---"

type Config struct {
	HydraAddr  string `envconfig:"TEST_HYDRA_ADDRESS"`
	User       string `envconfig:"TEST_USER_EMAIL"`
	Pwd        string `envconfig:"TEST_USER_PASSWORD"`
	ReqTimeout uint   `envconfig:"TEST_REQUEST_TIMEOUT,default=100"`
	ReqDelay   uint   `envconfig:"TEST_REQUEST_DELAY,default=5"`
	Domain     string `envconfig:"DOMAIN`
}

func TestApiGatewayIntegration(t *testing.T) {

	var conf Config
	err := envconfig.Init(&conf)
	if err != nil {
		panic(err)
	}

	httpClient, err := ingressgateway.FromEnv().Client()

	if err != nil {
		t.Fatalf("Unnable to initialize ingress gateway client: %v", err)
	}

	oauthClientID := generateRandomString(OauthClientIDLength)
	oauthClientSecret := generateRandomString(OauthClientSecretLength)

	oauth2Cfg := clientcredentials.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		TokenURL:     fmt.Sprintf("https://%s/oauth2/token", conf.HydraAddr),
		Scopes:       []string{"read"},
	}

	oauthClient := oauth2Cfg.Client(context.Background())
	tr := httpClient.Transport
	oauthClient.Transport = tr

	jwtConfig, err := jwt.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	tester := api.NewTester(httpClient, []retry.Option{
		retry.Delay(time.Duration(conf.ReqDelay) * time.Second),
		retry.Attempts(conf.ReqTimeout / conf.ReqDelay),
		retry.DelayType(retry.FixedDelay),
	})

	k8sClient := getDynamicClient()

	// create common resources for all scenarios
	globalCommonResources, err := manifestprocessor.ParseFromFileWithTemplate(globalCommonResourcesFile, manifestsDirectory, resourceSeparator, struct {
		OauthClientSecret string
		OauthClientID     string
	}{
		OauthClientSecret: base64.StdEncoding.EncodeToString([]byte(oauthClientSecret)),
		OauthClientID:     base64.StdEncoding.EncodeToString([]byte(oauthClientID)),
	})
	if err != nil {
		panic(err)
	}
	createResources(k8sClient, globalCommonResources...)
	time.Sleep(5 * time.Second)

	hydraClientResource, err := manifestprocessor.ParseFromFile(hydraClientFile, manifestsDirectory, resourceSeparator)
	if err != nil {
		panic(err)
	}
	createResources(k8sClient, hydraClientResource...)
	// defer deleting namespace (it will also delete all remaining resources in that namespace)
	defer func() {
		time.Sleep(time.Second * 3)
		nsResourceSchema, ns, name := getResourceSchemaAndNamespace(globalCommonResources[0])
		manager.DeleteResource(k8sClient, nsResourceSchema, ns, name)
	}()
	t.Run("parallel tests", func(t *testing.T) {
		t.Run("expose service without access strategy (plain access)", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, noAccessStrategyApiruleResource...)

			//for _, commonResource := range commonResources {
			//	resourceSchema, ns, name := getResourceSchemaAndNamespace(commonResource)
			//	manager.UpdateResource(k8sClient, resourceSchema, ns, name, commonResource)
			//}

			assert.NoError(t, tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			deleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose full service with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, resources...)

			token, err := oauth2Cfg.Token(context.Background())
			assert.Equal(t, err, nil)
			assert.NotNil(t, token)
			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), token.AccessToken))

			deleteResources(k8sClient, commonResources...)

			assert.NoError(t, tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose service with OAUTH and JWT on speficic paths", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			oauthStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "jwt-oauth", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, oauthStrategyApiruleResource...)

			tokenOAUTH, err := oauth2Cfg.Token(context.Background())
			assert.Nil(t, err)
			assert.NotNil(t, tokenOAUTH)

			tokenJWT, err := jwt.Authenticate(jwtConfig.IdProviderConfig)
			if err != nil {
				t.Fatalf("failed to fetch and id_token. %s", err.Error())
			}

			assert.Nil(t, err)
			assert.NotNil(t, tokenJWT)

			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.kyma.local/headers", testID), tokenOAUTH.AccessToken))
			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.kyma.local/image", testID), tokenJWT))

			deleteResources(k8sClient, commonResources...)

		})

		t.Run("Expose service with OAUTH and update to plain access ", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)

			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			resources, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "oauth2", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, resources...)

			token, err := oauth2Cfg.Token(context.Background())
			assert.Equal(t, err, nil)
			assert.NotNil(t, token)
			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), token.AccessToken))

			//Update API to give plain access
			namePrefix := strings.TrimSuffix(resources[0].GetName(), "-"+testID)

			unsecuredApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			updateResources(k8sClient, unsecuredApiruleResource...)

			assert.NoError(t, tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			deleteResources(k8sClient, commonResources...)

			assert.NoError(t, tester.TestDeletedAPI(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 strategy", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)
			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(t, tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(oauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			updateResources(k8sClient, securedApiruleResource...)

			token, err := oauth2Cfg.Token(context.Background())
			assert.Equal(t, err, nil)
			assert.NotNil(t, token)
			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain), token.AccessToken))

			deleteResources(k8sClient, commonResources...)
		})

		t.Run("Expose unsecured API next secure it with OAUTH2 and JWT strategy on paths", func(t *testing.T) {
			t.Parallel()
			testID := generateRandomString(testIDLength)
			// create common resources from files
			commonResources, err := manifestprocessor.ParseFromFileWithTemplate(testingAppFile, manifestsDirectory, resourceSeparator, struct{ TestID string }{TestID: testID})
			if err != nil {
				t.Fatalf("failed to process common manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, commonResources...)

			// create api-rule from file
			noAccessStrategyApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(noAccessStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: "unsecured", TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}
			createResources(k8sClient, noAccessStrategyApiruleResource...)

			assert.NoError(t, tester.TestUnsecuredEndpoint(fmt.Sprintf("https://httpbin-%s.%s", testID, conf.Domain)))

			//update to secure API

			namePrefix := strings.TrimSuffix(noAccessStrategyApiruleResource[0].GetName(), "-"+testID)

			securedApiruleResource, err := manifestprocessor.ParseFromFileWithTemplate(jwtAndOauthStrategyApiruleFile, manifestsDirectory, resourceSeparator, struct {
				NamePrefix string
				TestID     string
				Domain     string
			}{NamePrefix: namePrefix, TestID: testID, Domain: conf.Domain})
			if err != nil {
				t.Fatalf("failed to process resource manifest files for test %s, details %s", t.Name(), err.Error())
			}

			updateResources(k8sClient, securedApiruleResource...)

			oauth, err := oauth2Cfg.Token(context.Background())
			tokenJWT, err := jwt.Authenticate(jwtConfig.IdProviderConfig)
			if err != nil {
				t.Fatalf("failed to fetch and id_token. %s", err.Error())
			}

			assert.Nil(t, err)
			assert.NotNil(t, tokenJWT)

			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.kyma.local/headers", testID), oauth.AccessToken))
			assert.NoError(t, tester.TestSecuredEndpoint(fmt.Sprintf("https://httpbin-%s.kyma.local/image", testID), tokenJWT))
			deleteResources(k8sClient, commonResources...)
		})

	})
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return cfg
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	return cfg
}

func getDynamicClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(loadKubeConfigOrDie())
	if err != nil {
		panic(err)
	}
	return client
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyz")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func getResourceSchemaAndNamespace(manifest unstructured.Unstructured) (schema.GroupVersionResource, string, string) {
	metadata := manifest.Object["metadata"].(map[string]interface{})
	apiVersion := strings.Split(fmt.Sprintf("%s", manifest.Object["apiVersion"]), "/")
	namespace := "default"
	if metadata["namespace"] != nil {
		namespace = fmt.Sprintf("%s", metadata["namespace"])
	}
	resourceName := fmt.Sprintf("%s", metadata["name"])
	resourceKind := fmt.Sprintf("%s", manifest.Object["kind"])
	if resourceKind == "Namespace" {
		namespace = ""
	}
	//TODO: Move this ^ somewhere else and make it clearer
	apiGroup, version := getGroupAndVersion(apiVersion)
	resourceSchema := schema.GroupVersionResource{Group: apiGroup, Version: version, Resource: pluralForm(resourceKind)}
	return resourceSchema, namespace, resourceName
}

func createResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, _ := getResourceSchemaAndNamespace(resource)
		manager.CreateResource(k8sClient, resourceSchema, ns, resource)
	}
}

func updateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, _ := getResourceSchemaAndNamespace(resource)
		manager.UpdateResource(k8sClient, resourceSchema, ns, resource.GetName(), resource)
	}
}

func deleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, resource := range resources {
		resourceSchema, ns, name := getResourceSchemaAndNamespace(resource)
		manager.DeleteResource(k8sClient, resourceSchema, ns, name)
	}
}

func getGroupAndVersion(apiVersion []string) (apiGroup string, version string) {
	if len(apiVersion) > 1 {
		return apiVersion[0], apiVersion[1]
	}
	return "", apiVersion[0]
}

func pluralForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}
