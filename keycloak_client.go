package keycloak

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman.v2/plugin"
	"gopkg.in/h2non/gentleman.v2/plugins/query"
	"gopkg.in/h2non/gentleman.v2/plugins/timeout"
)

// Client is the keycloak client.
type Client struct {
	apiURL     *url.URL
	httpClient *gentleman.Client
}

// NewClient returns a keycloak client.
func NewClient(config Config) (*Client, error) {
	var uAPI *url.URL
	{
		var err error
		uAPI, err = url.Parse(config.AddrAPI)
		if err != nil {
			return nil, errors.Wrap(err, MsgErrCannotParse+"."+APIURL)
		}
	}

	var httpClient = gentleman.New()
	{
		httpClient = httpClient.URL(uAPI.String())
		httpClient = httpClient.Use(timeout.Request(config.Timeout))
	}

	var client = &Client{
		apiURL:     uAPI,
		httpClient: httpClient,
	}

	return client, nil
}

// GetToken returns a valid token from keycloak
func (c *Client) GetToken(realm string, username string, password string) (string, error) {
	var req *gentleman.Request
	{
		var authPath = fmt.Sprintf("/auth/realms/%s/protocol/openid-connect/token", realm)
		req = c.httpClient.Post()
		req = req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		req = req.Path(authPath)
		req = req.Type("urlencoded")
		req = req.BodyString(fmt.Sprintf("username=%s&password=%s&grant_type=password&client_id=admin-cli", username, password))
	}

	var resp *gentleman.Response
	{
		var err error
		resp, err = req.Do()
		if err != nil {
			return "", errors.Wrap(err, MsgErrCannotObtain+"."+TokenMsg)
		}
	}
	defer resp.Close()

	var unmarshalledBody map[string]interface{}
	{
		var err error
		err = resp.JSON(&unmarshalledBody)
		if err != nil {
			return "", errors.Wrap(err, MsgErrCannotUnmarshal+"."+Response)
		}
	}

	var accessToken interface{}
	{
		var ok bool
		accessToken, ok = unmarshalledBody["access_token"]
		if !ok {
			return "", fmt.Errorf(MsgErrMissingParam + "." + AccessToken)
		}
	}

	return accessToken.(string), nil
}

// get is a HTTP get method.
func (c *Client) get(accessToken string, data interface{}, plugins ...plugin.Plugin) error {
	var err error
	var req = c.httpClient.Get()
	req = applyPlugins(req, plugins...)
	req = setAuthorisationHeader(req, accessToken)

	if err != nil {
		return err
	}

	var resp *gentleman.Response
	{
		var err error
		resp, err = req.Do()
		if err != nil {
			return errors.Wrap(err, MsgErrCannotObtain+"."+Response)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return HTTPError{
				HTTPStatus: resp.StatusCode,
				Message:    string(resp.Bytes()),
			}
		}

		switch resp.Header.Get("Content-Type") {
		case "application/json":
			return resp.JSON(data)
		case "application/octet-stream":
			_ = resp.Bytes()
			return nil
		default:
			return fmt.Errorf("%s.%v", MsgErrUnkownHTTPContentType, resp.Header.Get("Content-Type"))
		}
	}
}

func (c *Client) post(accessToken string, data interface{}, plugins ...plugin.Plugin) (string, error) {
	var err error
	var req = c.httpClient.Post()
	req = applyPlugins(req, plugins...)
	req = setAuthorisationHeader(req, accessToken)

	if err != nil {
		return "", err
	}

	var resp *gentleman.Response
	{
		var err error
		resp, err = req.Do()
		if err != nil {
			return "", errors.Wrap(err, MsgErrCannotObtain+"."+Response)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return "", HTTPError{
				HTTPStatus: resp.StatusCode,
				Message:    string(resp.Bytes()),
			}
		}

		var location = resp.Header.Get("Location")

		switch resp.Header.Get("Content-Type") {
		case "application/json":
			return location, resp.JSON(data)
		case "application/octet-stream":
			data = resp.Bytes()
			return location, nil
		default:
			return location, nil
		}
	}
}

func (c *Client) delete(accessToken string, plugins ...plugin.Plugin) error {
	var err error
	var req = c.httpClient.Delete()
	req = applyPlugins(req, plugins...)
	req = setAuthorisationHeader(req, accessToken)

	if err != nil {
		return err
	}

	var resp *gentleman.Response
	{
		var err error
		resp, err = req.Do()
		if err != nil {
			return errors.Wrap(err, MsgErrCannotObtain+"."+Response)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return HTTPError{
				HTTPStatus: resp.StatusCode,
				Message:    string(resp.Bytes()),
			}
		}

		return nil
	}
}

func (c *Client) put(accessToken string, plugins ...plugin.Plugin) error {
	var err error
	var req = c.httpClient.Put()
	req = applyPlugins(req, plugins...)
	req = setAuthorisationHeader(req, accessToken)

	if err != nil {
		return err
	}

	var resp *gentleman.Response
	{
		var err error
		resp, err = req.Do()
		if err != nil {
			return errors.Wrap(err, MsgErrCannotObtain+"."+Response)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return HTTPError{
				HTTPStatus: resp.StatusCode,
				Message:    string(resp.Bytes()),
			}
		}

		return nil
	}
}

func setAuthorisationHeader(req *gentleman.Request, accessToken string) *gentleman.Request {
	var r = req.SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	r = r.SetHeader("X-Forwarded-Proto", "https")
	return r
}

// applyPlugins apply all the plugins to the request req.
func applyPlugins(req *gentleman.Request, plugins ...plugin.Plugin) *gentleman.Request {
	var r = req
	for _, p := range plugins {
		r = r.Use(p)
	}
	return r
}

// createQueryPlugins create query parameters with the key values paramKV.
func createQueryPlugins(paramKV ...string) []plugin.Plugin {
	var plugins = []plugin.Plugin{}
	for i := 0; i < len(paramKV); i += 2 {
		var k = paramKV[i]
		var v = paramKV[i+1]
		plugins = append(plugins, query.Add(k, v))
	}
	return plugins
}
