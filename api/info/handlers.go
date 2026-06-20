package info

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sentinel-official/sentinel-go-sdk/amneziawg"
	"github.com/sentinel-official/sentinel-go-sdk/hysteria2"
	"github.com/sentinel-official/sentinel-go-sdk/libs/geoip"
	"github.com/sentinel-official/sentinel-go-sdk/node"
	"github.com/sentinel-official/sentinel-go-sdk/openvpn"
	"github.com/sentinel-official/sentinel-go-sdk/types"
	"github.com/sentinel-official/sentinel-go-sdk/v2ray"
	"github.com/sentinel-official/sentinel-go-sdk/version"
	"github.com/sentinel-official/sentinel-go-sdk/wireguard"
	"github.com/sentinel-official/sentinel-go-sdk/xray"

	"github.com/sentinel-official/sentinel-dvpnx/core"
)

// serviceMetadata builds the redacted, client-facing metadata for one service.
func serviceMetadata(service types.ServerService) (any, error) {
	switch service.Type() { //nolint:exhaustive // default branch errors on unlisted service types
	case types.ServiceTypeOpenVPN:
		items, ok := service.Metadata().([]*openvpn.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement openvpn.ServerMetadata")
		}

		var md []*openvpn.ServerMetadata
		for _, v := range items {
			md = append(md, &openvpn.ServerMetadata{
				Protocol: v.Protocol,
			})
		}

		return md, nil
	case types.ServiceTypeV2Ray:
		items, ok := service.Metadata().([]*v2ray.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement v2ray.ServerMetadata")
		}

		var md []*v2ray.ServerMetadata
		for _, v := range items {
			md = append(md, &v2ray.ServerMetadata{
				ProxyProtocol:     v.ProxyProtocol,
				TransportProtocol: v.TransportProtocol,
				TransportSecurity: v.TransportSecurity,
			})
		}

		return md, nil
	case types.ServiceTypeWireGuard:
		items, ok := service.Metadata().([]*wireguard.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement wireguard.ServerMetadata")
		}

		var md []*wireguard.ServerMetadata
		for range items {
			md = append(md, &wireguard.ServerMetadata{})
		}

		return md, nil
	case types.ServiceTypeAmneziaWG:
		items, ok := service.Metadata().([]*amneziawg.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement amneziawg.ServerMetadata")
		}

		var md []*amneziawg.ServerMetadata
		for range items {
			md = append(md, &amneziawg.ServerMetadata{})
		}

		return md, nil
	case types.ServiceTypeHysteria2:
		items, ok := service.Metadata().([]*hysteria2.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement hysteria2.ServerMetadata")
		}

		var md []*hysteria2.ServerMetadata

		for _, v := range items {
			obfsPassword := v.ObfsPassword
			if obfsPassword != "" {
				obfsPassword = "<redacted>"
			}

			md = append(md, &hysteria2.ServerMetadata{
				ObfsPassword: obfsPassword,
			})
		}

		return md, nil
	case types.ServiceTypeXray:
		items, ok := service.Metadata().([]*xray.ServerMetadata)
		if !ok {
			return nil, errors.New("metadata does not implement xray.ServerMetadata")
		}

		var md []*xray.ServerMetadata
		for _, v := range items {
			md = append(md, &xray.ServerMetadata{
				ProxyProtocol:     v.ProxyProtocol,
				TransportProtocol: v.TransportProtocol,
				TransportSecurity: v.TransportSecurity,
				Flow:              v.Flow,
				Method:            v.Method,
			})
		}

		return md, nil
	default:
		return nil, errors.New("unknown service type")
	}
}

// buildServiceInfos builds one ServiceInfo per active service, ordered by service
// type, and returns the total peer count summed across all services.
func buildServiceInfos(c *core.Context) ([]node.ServiceInfo, int, error) {
	serviceTypes := c.ServiceTypes()

	infos := make([]node.ServiceInfo, 0, len(serviceTypes))
	total := 0

	for _, t := range serviceTypes {
		service, ok := c.Service(t)
		if !ok {
			continue
		}

		md, err := serviceMetadata(service)
		if err != nil {
			return nil, 0, fmt.Errorf("building metadata for service %q: %w", t, err)
		}

		peers := service.PeersLen()
		total += peers

		infos = append(infos, node.ServiceInfo{
			Type:     t.String(),
			Metadata: md,
			Peers:    peers,
		})
	}

	return infos, total, nil
}

// handlerGetInfo returns a handler function to retrieve node information.
func handlerGetInfo(c *core.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		infos, total, err := buildServiceInfos(c)
		if err != nil {
			err = fmt.Errorf("parsing metadata: %w", err)
			ctx.JSON(http.StatusInternalServerError, types.NewResponseError(1, err))

			return
		}

		dlSpeed, ulSpeed := c.SpeedtestResults()
		loc := c.Location()

		// Construct the result structure with node information.
		res := &node.GetInfoResult{
			Addr:         c.NodeAddr().String(),
			Downlink:     ulSpeed.String(),
			HandshakeDNS: c.HandshakeDNS(),
			Location: &geoip.Location{
				City:        loc.City,
				Country:     loc.Country,
				CountryCode: loc.CountryCode,
				Latitude:    loc.Latitude,
				Longitude:   loc.Longitude,
			},
			Moniker:  c.Moniker(),
			Peers:    total,
			Services: infos,
			Uplink:   dlSpeed.String(),
			Version:  version.Get(),
		}

		// Send the result as a JSON response with HTTP status 200.
		ctx.JSON(http.StatusOK, types.NewResponseResult(res))
	}
}
