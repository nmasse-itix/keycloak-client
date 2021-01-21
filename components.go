package keycloak

import (
	"gopkg.in/h2non/gentleman.v2/plugins/body"
	"gopkg.in/h2non/gentleman.v2/plugins/url"
)

const (
	componentsPath     = "/auth/admin/realms/:realm/components"
	componentsByIDPath = componentsPath + "/:id"
)

// GetComponents gets all components of the realm
func (c *Client) GetComponents(accessToken string, realmName string) ([]ComponentRepresentation, error) {
	var resp = []ComponentRepresentation{}
	var err = c.get(accessToken, &resp, url.Path(componentsPath), url.Param("realm", realmName))
	return resp, err
}

// GetComponent gets a component of the realm
func (c *Client) GetComponent(accessToken string, realmName string, componentID string) ([]ComponentRepresentation, error) {
	var resp = []ComponentRepresentation{}
	var err = c.get(accessToken, &resp, url.Path(componentsByIDPath), url.Param("realm", realmName), url.Param("id", componentID))
	return resp, err
}

// CreateComponent creates a new component in the realm
func (c *Client) CreateComponent(accessToken string, realmName string, component ComponentRepresentation) (string, error) {
	return c.post(accessToken, nil, url.Path(componentsPath), url.Param("realm", realmName), body.JSON(component))
}

// UpdateComponent updates a new component in the realm
func (c *Client) UpdateComponent(accessToken string, realmName, componentID string, component ComponentRepresentation) error {
	return c.put(accessToken, nil, url.Path(componentsPath), url.Param("realm", realmName), url.Param("id", componentID), body.JSON(component))
}

// DeleteComponent deletes a component in the realm
func (c *Client) DeleteComponent(accessToken string, realmName, componentID string) error {
	return c.delete(accessToken, nil, url.Path(componentsPath), url.Param("realm", realmName), url.Param("id", componentID))
}
