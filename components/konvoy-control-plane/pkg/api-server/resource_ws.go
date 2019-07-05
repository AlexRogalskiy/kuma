package api_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log" // todo(jakubdyszkiewicz) replace with core
)

const namespace = "default"

type ResourceWsDefinition struct {
	Name                string
	Path                string
	ResourceFactory     func() model.Resource
	ResourceListFactory func() model.ResourceList
}

type resourceWs struct {
	resourceStore store.ResourceStore
	readOnly      bool
	ResourceWsDefinition
}

type ResourceReqResp struct {
	Type string             `json:"type"`
	Name string             `json:"name"`
	Mesh string             `json:"mesh"`
	Spec model.ResourceSpec `json:"-"`
}

var (
	marshaler   = jsonpb.Marshaler{}
	unmarshaler = jsonpb.Unmarshaler{AllowUnknownFields: true}
)

// To marshall spec we have to use package from jsonb that handle edge cases of protobuf -> json conversion
func (r ResourceReqResp) MarshalJSON() ([]byte, error) {
	type tmp ResourceReqResp // otherwise we've got stackoverflow
	res := tmp(r)
	var specJsonBytes bytes.Buffer
	if err := marshaler.Marshal(&specJsonBytes, res.Spec); err != nil {
		return nil, err
	}
	metaJsonBytes, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	// join both jsons
	data := make(map[string]interface{})
	if err := json.Unmarshal(specJsonBytes.Bytes(), &data); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(metaJsonBytes, &data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (r *ResourceReqResp) UnmarshalJSON(jsonBytes []byte) error {
	if err := unmarshaler.Unmarshal(bytes.NewBuffer(jsonBytes), r.Spec); err != nil {
		return err
	}
	// json.Unmarshal cannot ignore unknown fields therefore we have to unmarshall to map
	values := map[string]string{}
	if err := json.Unmarshal(jsonBytes, &values); err != nil {
		return err
	}
	r.Name = values["name"]
	r.Type = values["type"]
	r.Mesh = values["mesh"]
	return nil
}

func (r *resourceWs) NewWs() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path(fmt.Sprintf("/meshes/{mesh}/%s", r.Path)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Param(ws.PathParameter("mesh", "Name of the Mesh").DataType("string"))

	ws.Route(ws.GET("/{name}").To(r.findResource).
		Doc(fmt.Sprintf("Get a %s", r.Name)).
		Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
		//Writes(r.SpecFactory()).
		Returns(200, "OK", nil). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
		Returns(404, "Not found", nil))

	ws.Route(ws.GET("").To(r.listResources).
		Doc(fmt.Sprintf("List of %s", r.Name)).
		//Writes(r.SampleListSpec).
		Returns(200, "OK", nil)) // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp

	if !r.readOnly {
		ws.Route(ws.PUT("/{name}").To(r.createOrUpdateResource).
			Doc(fmt.Sprintf("Updates a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of the %s", r.Name)).DataType("string")).
			//Reads(r.SampleSpec). // todo(jakubdyszkiewicz) figure out how to expose the doc for ResourceReqResp
			Returns(200, "OK", nil).
			Returns(201, "Created", nil))

		ws.Route(ws.DELETE("/{name}").To(r.deleteResource).
			Doc(fmt.Sprintf("Deletes a %s", r.Name)).
			Param(ws.PathParameter("name", fmt.Sprintf("Name of a %s", r.Name)).DataType("string")).
			Returns(200, "OK", nil))
	}

	return ws
}

func (r *resourceWs) findResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")

	// todo(jakubdyszkiewicz) find by mesh?
	resource := r.ResourceFactory()
	err := r.resourceStore.Get(request.Request.Context(), resource, store.GetByName(namespace, name))
	if err != nil {
		if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
			writeError(response, 404, "")
		} else {
			log.Log.Error(err, "Could not retrieve a resource", "name", name)
			writeError(response, 500, "Could not retrieve a resource")
		}
	} else {
		res := &ResourceReqResp{
			Type: string(resource.GetType()),
			Name: name,
			Mesh: meshName,
			Spec: resource.GetSpec(),
		}
		err = response.WriteAsJson(res)
		if err != nil {
			log.Log.Error(err, "Could not write the response")
		}
	}
}

type resourceSpecList struct {
	Items []*ResourceReqResp `json:"items"`
}

func (r *resourceWs) listResources(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")

	list := r.ResourceListFactory()
	// todo(jakubdyszkiewicz) find by mesh?
	if err := r.resourceStore.List(request.Request.Context(), list, store.ListByNamespace(namespace)); err != nil {
		log.Log.Error(err, "Could not retrieve resources")
		writeError(response, 500, "Could not list a resource")
	} else {
		var items []*ResourceReqResp
		for _, item := range list.GetItems() {
			items = append(items, &ResourceReqResp{
				Type: string(item.GetType()),
				Name: item.GetMeta().GetName(),
				Mesh: meshName,
				Spec: item.GetSpec(),
			})
		}
		specList := resourceSpecList{Items: items}
		if err := response.WriteAsJson(specList); err != nil {
			log.Log.Error(err, "Could not write as JSON", "type", string(list.GetItemType()))
			writeError(response, 500, "Could not list a resource")
		}
	}
}

func (r *resourceWs) createOrUpdateResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	resourceRes := ResourceReqResp{
		Spec: r.ResourceFactory().GetSpec(),
	}

	err := request.ReadEntity(&resourceRes)
	if err != nil {
		log.Log.Error(err, "Could not read an entity")
		writeError(response, 400, "Could not process the resource")
	}

	if err := r.validateResourceRequest(request, &resourceRes); err != nil {
		writeError(response, 400, err.Error())
	} else {
		resource := r.ResourceFactory()
		// todo(jakubdyszkiewicz) find by mesh?
		if err := r.resourceStore.Get(request.Request.Context(), resource, store.GetByName(namespace, name)); err != nil {
			if err.Error() == store.ErrorResourceNotFound(resource.GetType(), namespace, name).Error() {
				r.createResource(request.Request.Context(), name, resourceRes.Spec, response)
			} else {
				log.Log.Error(err, "Could get a resource from the store", "namespace", namespace, "name", name, "type", string(resource.GetType()))
				writeError(response, 500, "Could not create a resource")
			}
		} else {
			r.updateResource(request.Request.Context(), resource, resourceRes.Spec, response)
		}
	}
}

func (r *resourceWs) validateResourceRequest(request *restful.Request, resourceReq *ResourceReqResp) error {
	name := request.PathParameter("name")
	meshName := request.PathParameter("mesh")
	if name != resourceReq.Name {
		return errors.New("Name from the URL has to be the same as in body")
	}
	if string(r.ResourceFactory().GetType()) != resourceReq.Type {
		return errors.New("Type from the URL has to be the same as in body")
	}
	if meshName != resourceReq.Mesh {
		return errors.New("Mesh from the URL has to be the same as in body")
	}
	return nil
}

func (r *resourceWs) createResource(ctx context.Context, name string, spec model.ResourceSpec, response *restful.Response) {
	res := r.ResourceFactory()
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Create(ctx, res, store.CreateByName(namespace, name)); err != nil {
		log.Log.Error(err, "Could not create a resource")
		writeError(response, 500, "Could not create a resource")
	} else {
		response.WriteHeader(201)
	}
}

func (r *resourceWs) updateResource(ctx context.Context, res model.Resource, spec model.ResourceSpec, response *restful.Response) {
	_ = res.SetSpec(spec)
	if err := r.resourceStore.Update(ctx, res); err != nil {
		log.Log.Error(err, "Could not update a resource")
		writeError(response, 500, "Could not update a resource")
	} else {
		response.WriteHeader(200)
	}
}

func (r *resourceWs) deleteResource(request *restful.Request, response *restful.Response) {
	name := request.PathParameter("name")

	resource := r.ResourceFactory()
	// todo(jakubdyszkiewicz) delete by mesh?
	err := r.resourceStore.Delete(request.Request.Context(), resource, store.DeleteByName(namespace, name))
	if err != nil {
		writeError(response, 500, "Could not delete a resource")
		log.Log.Error(err, "Could not delete a resource", "namespace", namespace, "name", name, "type", string(resource.GetType()))
	}
}

func writeError(response *restful.Response, httpStatus int, msg string) {
	if err := response.WriteErrorString(httpStatus, msg); err != nil {
		log.Log.Error(err, "Cloud not write the response")
	}
}
