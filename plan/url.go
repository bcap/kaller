package plan

import (
	"encoding/json"
	"net/url"

	"gopkg.in/yaml.v3"
)

// The URL type is in essence the same as net/url.URL, but with some facilities for
// cleaner json and yaml serialization/deserialization
type URL struct {
	*url.URL
}

func (u URL) MarshalYAML() (interface{}, error) {
	return u.URL.String(), nil
}

func (u *URL) UnmarshalYAML(node *yaml.Node) error {
	urlStruct, err := url.Parse(node.Value)
	if err != nil {
		return err
	}
	u.URL = urlStruct
	return nil
}

func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.URL.String())
}

func (u *URL) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	urlStruct, err := url.Parse(str)
	if err != nil {
		return err
	}
	u.URL = urlStruct
	return nil
}

func MustParseURL(u string) URL {
	urlStruct, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return URL{URL: urlStruct}
}
