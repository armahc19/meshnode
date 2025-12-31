package main

import (
	"context"
//	"encoding/hex"
	"fmt"
	"log"
//	"os/exec"
	"strings"
	"time"
	"regexp"

	cid "github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	dht "github.com/libp2p/go-libp2p-kad-dht"
)

// ServiceID is the unique hash of the application image (the "What")
type ServiceID string

// ServiceDefinition defines the requirements and identity of an app
type ServiceDefinition struct {
	ID           ServiceID `json:"id"`           // The Image Hash/Digest
	Name         string    `json:"name"`         // Human readable name
	InternalPort int       `json:"internal_port"` // Port the app listens on inside container
	Protocol     string    `json:"protocol"`      // "http", "tcp", etc.
}

// ServiceInstance defines a LIVE running service on the mesh (the "Where")
type ServiceInstance struct {
	Definition ServiceDefinition `json:"definition"`
	HostPeerID string            `json:"host_peer_id"` // The PeerID of the node running it
	MeshPort   int               `json:"mesh_port"`    // The dynamic port assigned by the host
	Status     string            `json:"status"`       // "starting", "running", "degraded"
}

// FullIdentity returns a string representation for logging
func (s *ServiceInstance) FullIdentity() string {
	return fmt.Sprintf("%s:%d (on %s)", s.Definition.ID, s.MeshPort, s.HostPeerID)
}

func CreateDefinitionFromRuntime(info RuntimeInfo, imageHash string, port int) ServiceDefinition {
	return ServiceDefinition{
		ID:           ServiceID(imageHash),
		Name:         info.Runtime + "-app",
		InternalPort: port,
		Protocol:     "http", // Defaulting to HTTP for MVP
	}
}

// GetImageHash fetches the unique SHA256 ID of a local Docker image
func GetContainerNameAsID(containerName string) (ServiceID, error) {
    // Validate container name format
    containerName = strings.ToLower(strings.TrimSpace(containerName))
    
    // Ensure it's DNS-safe
    if !isValidDNSLabel(containerName) {
        return "", fmt.Errorf("container name '%s' is not DNS-safe", containerName)
    }
    
    return ServiceID(containerName), nil
}

func isValidDNSLabel(name string) bool {
    // DNS labels: 1-63 chars, a-z, 0-9, hyphen, not start/end with hyphen
    if len(name) < 1 || len(name) > 63 {
        return false
    }
    
    // Regex for DNS label
    matched, _ := regexp.MatchString(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`, name)
    return matched
}


func RegisterLocalService(hID string, containerName string, externalPort int) (*ServiceInstance, error) {
    // 1. Validate and create ServiceID from container name
    serviceID, err := GetContainerNameAsID(containerName)
    if err != nil {
        return nil, err
    }

    // 2. Define the Service
    def := ServiceDefinition{
        ID:           serviceID,
        Name:         containerName, // Human readable name = container name
        InternalPort: 80,            // Default container port
        Protocol:     "http",
    }

    // 3. Create the Instance
    instance := ServiceInstance{
        Definition: def,
        HostPeerID: hID,
        MeshPort:   externalPort,
        Status:     "running",
    }

    // 4. Save to local state
    MyHostedServices[serviceID] = instance
    
    fmt.Printf("✔ Registered Service: %s (Port: %d)\n", containerName, externalPort)
    return &instance, nil
}


// ServiceIDToCID converts our hex image hash into a libp2p-compatible CID
// ServiceIDToCID - Keep but simplify (no more hex decoding)
func ServiceIDToCID(id ServiceID) (cid.Cid, error) {
    // Convert container name directly to multihash
    mhash, err := mh.Sum([]byte(id), mh.SHA2_256, -1)
    if err != nil {
        return cid.Undef, fmt.Errorf("multihash encode failed: %v", err)
    }
    
    // Create CID V1
    return cid.NewCidV1(cid.Raw, mhash), nil
}

func StartServiceAdvertisement(ctx context.Context, kadDHT *dht.IpfsDHT) {
    ticker := time.NewTicker(5 * time.Minute) // Advertise every 5 mins
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // Loop through all services this node is currently hosting
            for id, instance := range MyHostedServices {
                c, err := ServiceIDToCID(id)
                if err != nil {
                    log.Printf("❌ Failed to create CID for service %s: %v", id, err)
                    continue
                }

                // Announce to the DHT that this node provides this Service CID
                // 'true' means we want to broadcast this to the network
                if err := kadDHT.Provide(ctx, c, true); err != nil {
                    log.Printf("❌ DHT: Failed to advertise service %s: %v", id, err)
                } else {
                    log.Printf("📢 DHT: Advertised service instance: %s", instance.FullIdentity())
                }
            }
        }
    }
}
