package api_server

import (
	"os"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/api-server/types"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

var APIIndexResponseFn = kumaAPIIndexResponse

func addIndexWsEndpoints(ws *restful.WebService, getInstanceId func() string, getClusterId func() string) error {
	hostname, err := os.Hostname()
	var instanceId string
	var clusterId string
	if err != nil {
		return err
	}
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		if instanceId == "" {
			instanceId = getInstanceId()
		}

		if clusterId == "" {
			clusterId = getClusterId()
		}

		response := APIIndexResponseFn(hostname, instanceId, clusterId)

		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return nil
}

func kumaAPIIndexResponse(hostname string, instanceId string, clusterId string) interface{} {
	return types.IndexResponse{
		Hostname:   hostname,
		Tagline:    kuma_version.Product,
		Version:    kuma_version.Build.Version,
		InstanceId: instanceId,
		ClusterId:  clusterId,
	}
}
