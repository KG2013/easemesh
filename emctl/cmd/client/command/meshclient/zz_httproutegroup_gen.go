/*
Copyright (c) 2021, MegaEase
All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// code generated by github.com/megaease/easemeshctl/cmd/generator, DO NOT EDIT.
package meshclient

import (
	"context"
	"encoding/json"
	"fmt"
	v2alpha1 "github.com/megaease/easemesh-api/v2alpha1"
	resource "github.com/megaease/easemeshctl/cmd/client/resource"
	client "github.com/megaease/easemeshctl/cmd/common/client"
	errors "github.com/pkg/errors"
	"net/http"
)

type hTTPRouteGroupGetter struct {
	client *meshClient
}
type hTTPRouteGroupInterface struct {
	client *meshClient
}

func (h *hTTPRouteGroupGetter) HTTPRouteGroup() HTTPRouteGroupInterface {
	return &hTTPRouteGroupInterface{client: h.client}
}
func (h *hTTPRouteGroupInterface) Get(args0 context.Context, args1 string) (*resource.HTTPRouteGroup, error) {
	url := fmt.Sprintf("http://"+h.client.server+apiURL+"/mesh/"+"httproutegroups/%s", args1)
	r0, err := client.NewHTTPJSON().GetByContext(args0, url, nil, nil).HandleResponse(func(buff []byte, statusCode int) (interface{}, error) {
		if statusCode == http.StatusNotFound {
			return nil, errors.Wrapf(NotFoundError, "get HTTPRouteGroup %s", args1)
		}
		if statusCode >= 300 {
			return nil, errors.Errorf("call %s failed, return status code %d text %+v", url, statusCode, string(buff))
		}
		HTTPRouteGroup := &v2alpha1.HTTPRouteGroup{}
		err := json.Unmarshal(buff, HTTPRouteGroup)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal data to v2alpha1.HTTPRouteGroup")
		}
		return resource.ToHTTPRouteGroup(HTTPRouteGroup), nil
	})
	if err != nil {
		return nil, err
	}
	return r0.(*resource.HTTPRouteGroup), nil
}
func (h *hTTPRouteGroupInterface) Patch(args0 context.Context, args1 *resource.HTTPRouteGroup) error {
	url := fmt.Sprintf("http://"+h.client.server+apiURL+"/mesh/"+"httproutegroups/%s", args1.Name())
	object := args1.ToV2Alpha1()
	_, err := client.NewHTTPJSON().PutByContext(args0, url, object, nil).HandleResponse(func(b []byte, statusCode int) (interface{}, error) {
		if statusCode == http.StatusNotFound {
			return nil, errors.Wrapf(NotFoundError, "patch HTTPRouteGroup %s", args1.Name())
		}
		if statusCode < 300 && statusCode >= 200 {
			return nil, nil
		}
		return nil, errors.Errorf("call PUT %s failed, return statuscode %d text %+v", url, statusCode, string(b))
	})
	return err
}
func (h *hTTPRouteGroupInterface) Create(args0 context.Context, args1 *resource.HTTPRouteGroup) error {
	url := "http://" + h.client.server + apiURL + "/mesh/httproutegroups"
	object := args1.ToV2Alpha1()
	_, err := client.NewHTTPJSON().PostByContext(args0, url, object, nil).HandleResponse(func(b []byte, statusCode int) (interface{}, error) {
		if statusCode == http.StatusConflict {
			return nil, errors.Wrapf(ConflictError, "create HTTPRouteGroup %s", args1.Name())
		}
		if statusCode < 300 && statusCode >= 200 {
			return nil, nil
		}
		return nil, errors.Errorf("call Post %s failed, return statuscode %d text %+v", url, statusCode, string(b))
	})
	return err
}
func (h *hTTPRouteGroupInterface) Delete(args0 context.Context, args1 string) error {
	url := fmt.Sprintf("http://"+h.client.server+apiURL+"/mesh/"+"httproutegroups/%s", args1)
	_, err := client.NewHTTPJSON().DeleteByContext(args0, url, nil, nil).HandleResponse(func(b []byte, statusCode int) (interface{}, error) {
		if statusCode == http.StatusNotFound {
			return nil, errors.Wrapf(NotFoundError, "Delete HTTPRouteGroup %s", args1)
		}
		if statusCode < 300 && statusCode >= 200 {
			return nil, nil
		}
		return nil, errors.Errorf("call Delete %s failed, return statuscode %d text %+v", url, statusCode, string(b))
	})
	return err
}
func (h *hTTPRouteGroupInterface) List(args0 context.Context) ([]*resource.HTTPRouteGroup, error) {
	url := "http://" + h.client.server + apiURL + "/mesh/httproutegroups"
	result, err := client.NewHTTPJSON().GetByContext(args0, url, nil, nil).HandleResponse(func(b []byte, statusCode int) (interface{}, error) {
		if statusCode == http.StatusNotFound {
			return nil, errors.Wrapf(NotFoundError, "list service")
		}
		if statusCode >= 300 && statusCode < 200 {
			return nil, errors.Errorf("call GET %s failed, return statuscode %d text %+v", url, statusCode, b)
		}
		hTTPRouteGroup := []v2alpha1.HTTPRouteGroup{}
		err := json.Unmarshal(b, &hTTPRouteGroup)
		if err != nil {
			return nil, errors.Wrapf(err, "unmarshal data to v2alpha1.")
		}
		results := []*resource.HTTPRouteGroup{}
		for _, item := range hTTPRouteGroup {
			copy := item
			results = append(results, resource.ToHTTPRouteGroup(&copy))
		}
		return results, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]*resource.HTTPRouteGroup), nil
}