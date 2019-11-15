package sdk

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/logging"
)

// NewClient creates a new APIClient object
func NewClient(user, password, tenant, baseURL string, insecure bool) APIClient {

	t := http.DefaultTransport.(*http.Transport)
	t.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: insecure,
	}
	transport := logging.NewTransport("VRA7", t)
	httpClient := &http.Client{
		// Timeout:   clientTimeout,
		Transport: transport,
	}
	apiClient := APIClient{
		Username:    user,
		Password:    password,
		Tenant:      tenant,
		BaseURL:     baseURL,
		Insecure:    insecure,
		BearerToken: "",
		Client:      httpClient,
	}
	return apiClient
}

// DoRequest makes the request and returns the response
func (c *APIClient) DoRequest(req *APIRequest, login bool) (*APIResponse, error) {
	r, err := FromAPIRequestToHTTPRequest(req)
	if err != nil {
		return nil, err
	}
	if !login {
		c.Authenticate()
		r.Header.Add(AuthorizationHeader, c.BearerToken)
	}
	r.Header.Add(ConnectionHeader, CloseConnection)
	resp, err := c.Client.Do(r)
	if err != nil {
		log.Error("An error occurred when calling %v on %v. Error: %v", req.Method, req.URL, err)
		return nil, err
	}
	log.Info("Check the status of the request %s \n The response is: %s", req.URL, string(resp.Status))
	return FromHTTPRespToAPIResp(resp)
}

// Authenticate authenticates for the first time when the provider is invoked
func (c *APIClient) Authenticate() error {
	uri := fmt.Sprintf("%s"+Tokens, c.BaseURL)
	data := AuthenticationRequest{
		Username: c.Username,
		Password: c.Password,
		Tenant:   c.Tenant,
	}

	jsonData, _ := json.Marshal(data)

	req := &APIRequest{
		Method: POST,
		Body:   bytes.NewBufferString(string(jsonData)),
		URL:    uri,
	}
	req.AddHeader(AcceptHeader, AppJSON)
	req.AddHeader(ContentTypeHeader, AppJSON)

	return c.DoLogin(req)
}

// DoLogin returns the bearer token
func (c *APIClient) DoLogin(apiReq *APIRequest) error {
	apiResp, err := c.DoRequest(apiReq, true)
	if err != nil {
		return err
	}
	response := &AuthResponse{}

	err = json.Unmarshal(apiResp.Body, response)
	if err != nil {
		return err
	}
	c.BearerToken = fmt.Sprintf("Bearer %s", response.ID)
	return nil
}
