package strategy
import (
        "errors"
        "strings"

        "github.com/docker/swarm/cluster"
        "github.com/docker/swarm/scheduler/node"

        log "github.com/Sirupsen/logrus"
        "github.com/docker/swarm/scheduler/ibm/ego/egomanager"
)

const (
        // SwarmLabelNamespace defines the key prefix in all custom labels
        SwarmLabelNamespace = "com.docker.swarm"
        pkg = "strategy"
        class = "EGOStrategy"
)

var (
        ErrTenantIdNotProvided = errors.New("Tenant ID not provided.  Use --Label option to provide tenant id, e.g. docker run --Labal tenentid=<tenant id> ...")
        ErrEGONoResourcesAvailable = errors.New("no resources available to schedule container")
        ErrEGOInitFailed = errors.New("Creation of EGO manager failed.")
        egoManagerClient *egomanager.EGOManager
)

type EGOPlacementStrategy struct {
}


func (p *EGOPlacementStrategy) Initialize() error {
        var fn = "Initialize"
        em := egomanager.New()
        if em == nil {
                log.Debugf("[%s/%s/%s] %s",  pkg, class, fn, ErrEGOInitFailed.Error())
                return ErrEGOInitFailed
        }

        egoManagerClient = em
        err := egoManagerClient.Initialize()
        if err != nil {
                //error := err.Error()
                //if error != "operation now in progress" {
                if !strings.Contains(err.Error(), "Operation now in progress") {
                        log.Debugf("[%s/%s/%s] Error initializing EGO Manager:  %v",  pkg, class, fn, err)
                        return err
                }
        }
        return nil
}

func getTenantId(config *cluster.ContainerConfig) string {
        var tenantId string = ""
        if config.Labels != nil {
                tenantId = config.Labels["tenantid"]
        }
        return tenantId
}

func getContainerLabels(config *cluster.ContainerConfig) error {
        var (
                fn = "getContainerLabels"
        )
        log.Debugf("[%s/%s/%s] Entering...", pkg, class, fn)
        if config.Labels == nil {
                log.Debugf("[%s/%s/%s] No labels in this container.", pkg, class, fn)
                return nil
        }

        for key, value := range config.Labels {
                log.Debugf("[%s/%s/%s] Label[%v] = %v", pkg, class, fn, key, value)
        }
        return nil
}

func setEGOAllocIdLabel(config *cluster.ContainerConfig, allocid string) error {
        var (
                pkg = "strategy"
                class = "EGOStrategy"
                fn = "setEGOAllocIdLabel"
        )
        log.Debugf("[%s/%s/%s] Entering...", pkg, class, fn)
        if config.Labels == nil {
                config.Labels = make(map[string]string)
        }
        var key string = SwarmLabelNamespace+"ego.allocid"
        config.Labels[key] = allocid
        log.Debugf("[%s/%s/%s] Setting container label %s=%s.", pkg, class, fn, key, allocid)

        return nil
}
func reclaimAllocation(node *node.Node, reclaimAllocId string) error {
        var (
                pkg = "strategy"
                class = "EGOStrategy"
                fn = "reclaimAllocation"
        )
        log.Debugf("[%s/%s/%s] Entering...", pkg, class, fn)

        containers := node.Containers

        var key string = SwarmLabelNamespace+"ego.allocid"
        for _, c := range containers {
                if c != nil {
                        swarmID := c.Config.SwarmID()
                        if swarmID != "" {
                                memory := c.Config.Memory
                                cpus := c.Config.CpuShares
                                id := c.Config.Labels[key]
                                if id == reclaimAllocId {
                                        log.Debugf("[%s/%s/%s] Found containter with Label[com.docker.swarmego.allocid] = %s for swarm id: %v memory: %v, cpus: %v.", pkg, class, fn, id, swarmID, memory, cpus)
                                        e := c.Engine
                                        if e == nil {
                                                log.Debugf("[%s/%s/%s] No Engine found for container with swarm id: %v.", pkg, class, fn, swarmID)
                                        } else {
                                                cid := c.Id
                                                log.Debugf("[%s/%s/%s] Engine found for container id: %v.  Attempting to remove it.", pkg, class, fn, cid)
                                                err := e.RemoveContainer(c, true)
                                                if err != nil {
                                                        log.Debugf("[%s/%s/%s] Failed to remove container id: %v error: %v.", pkg, class, fn, cid, err.Error())
                                                } else {
                                                        log.Debugf("[%s/%s/%s] Successfully to removed container id: %v.", pkg, class, fn, cid)
                                                }
                                        }
                                }
                        }
                }
        }
        return nil
}
// Name returns the name of the strategy
func (p *EGOPlacementStrategy) Name() string {
        return "ego"
}

func (p *EGOPlacementStrategy) ReleaseAllocation(id string) error {
        err := egoManagerClient.Release(id)
        return err
}

func (p *EGOPlacementStrategy) PlaceContainer(config *cluster.ContainerConfig, nodes []*node.Node) (*node.Node, error) {

        var (
                pkg = "strategy"
                class = "EGOStrategy"
                fn = "PlaceContainer"
        )
        log.Debugf("[%s/%s/%s] Entering...", pkg, class, fn)

        var hostName string = ""
        var allocId string = ""
        var reclaimId string = ""

        var tenantId string = getTenantId(config)
        if tenantId == "" {
                log.Debugf("[%s/%s/%s] Tenant ID was not provided.", pkg, class, fn)
                return nil, ErrTenantIdNotProvided
        }
        /* EGO Client registration request object. */
        //rc, error := egoManagerClient.RegisterMD()
        //if rc < 0 {
        //      log.Debugf("[%s/%s/%s] Client registration request allocation failed.", pkg, class, fn)
        //      if error != nil {
        //              log.Debugf("[%s/%s/%s] Error code: %v", pkg, class, fn, error)
        //      }
        //} else {
        //      log.Debugf("[%s/%s/%s] EGO client name registration complete.  Return code: %v", pkg, class, fn, rc)

        //      /* Allocation request. */
        //      hostName, allocId, error = egoManagerClient.AllocMD(tenantId)
        //}

        /* Allocation request. */
        hostName, allocId, reclaimId, _ = egoManagerClient.AllocMD(tenantId)

        log.Debugf("[%s/%s/%s] Looking for match to ego host: %s.", pkg, class, fn, hostName)
        var placementHostName string
        if hostName != "" {
                setEGOAllocIdLabel(config, allocId)
                getContainerLabels(config)
                tokens := strings.Split(hostName, ".")
                for i := range tokens {
                        placementHostName = tokens[i]
                        break /* Just check the first token */
                }
        }

        if placementHostName == "" {
                log.Debugf("[%s/%s/%s] No match found for placement host node because name is empty.", pkg, class, fn)
                return nil, ErrEGONoResourcesAvailable
        }


        var matchNode *node.Node = nil

        for _, node := range nodes {
               if placementHostName == node.Name {
                       log.Debugf("[%s/%s/%s] Match found, node ID: %s, Name: %s .", pkg, class, fn, node.ID, node.Name)
                       matchNode = node
                       break
               } else {
                       log.Debugf("[%s/%s/%s] Node ID: %s, Name: %s is not a match.", pkg, class, fn, node.ID, node.Name)
               }
        }

        if matchNode == nil {
                log.Debugf("[%s/%s/%s] No match found for placement host node because name is empty.", pkg, class, fn)
                return nil, ErrEGONoResourcesAvailable
        } else {
                if reclaimId != "" {
                        reclaimAllocation(matchNode, reclaimId)
                }
        }
        return matchNode, nil
}

