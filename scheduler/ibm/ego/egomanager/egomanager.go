package egomanager
import (
        //"errors"
        //"strings"
        //"sync"
        "time"
        //"unsafe"

        //"github.com/docker/swarm/cluster"
        //"github.com/docker/swarm/scheduler/node"

        log "github.com/Sirupsen/logrus"
        "ego/sdk/goAPI/library/egomodule"
)

const (
        pkg   = "egomanager"
        class = "EGOManager"
        RESOURCE_ADD_MD = "resourceAddMD"
        RESOURCE_RECLAIM_MD = "resourceClaimMD"
)

var (
        handle *egomodule.GoEGOHandle
)

// EGOManager is exported
type EGOManager struct {
        //sync.Mutex
        egoModuleClient *egomodule.EGOModule
}

// New is exported
func New() *EGOManager {
        c := egomodule.NewEGOModule()
        return &EGOManager{
                egoModuleClient: c,
        }
}

func (m *EGOManager) Initialize() error {
        var fn = "Initialize"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)

        if m.egoModuleClient  == nil {
                log.Debugf("[%s/%s/%s] EGO module client is nil.",  pkg, class, fn)
        }

        h, error := m.egoModuleClient.GoEGOOpen()
        if error != nil {
                log.Debugf("[%s/%s/%s] EGO open error: %v",  pkg, class, fn, error)
        }
        handle = h
        if handle != nil {
                m.egoModuleClient.GoEGOLogon(handle)
                log.Debugf("[%s/%s/%s] EGO Logon completed.",  pkg, class, fn)

                rc, error := m.egoModuleClient.GoEGORegisterMD(handle)
                log.Debugf("[%s/%s/%s] EGO client name registration complete.  Return code: %v", pkg, class, fn, rc)
                if rc < 0 {
                        log.Debugf("[%s/%s/%s] Client registration failed.", pkg, class, fn)
                        if error != nil {
                                log.Debugf("[%s/%s/%s] Error code: %v", pkg, class, fn, error)
                                return error
                        }
                } else {
                        log.Debugf("[%s/%s/%s] Client registration successful.", pkg, class, fn)
                }
        } else {
                log.Debugf("[%s/%s/%s] EGO client handle is null.", pkg, class, fn)
        }
        return nil
}

func (m *EGOManager) Logon() error {
        var fn = "Logon"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)
        m.egoModuleClient.GoEGOLogon(handle)
        return nil
}


func (m *EGOManager) RegisterMD() (int, error) {
        var fn = "RegisterMD"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)

        if handle == nil {
                log.Debugf("[%s/%s/%s] EGO client handle is null.", pkg, class, fn)
                return -1, nil
        }

        if handle.EGO_handle == nil {
                log.Debugf("[%s/%s/%s] EGO handle is null.", pkg, class, fn)
                return -1, nil
        }

        rc, error := m.egoModuleClient.GoEGORegisterMD(handle)
        log.Debugf("[%s/%s/%s] EGO client name registration complete.  Return code: %v", pkg, class, fn, rc)
        if rc < 0 {
                log.Debugf("[%s/%s/%s] Client registration request allocation failed.", pkg, class, fn)
                if error != nil {
                        log.Debugf("[%s/%s/%s] Error code: %v", pkg, class, fn, error)
                        return rc, error
                }
        }
        return rc, nil
}

func (m *EGOManager) AllocMD(consumerId string) (string, string, string, error) {
        var fn = "AllocMD"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)

        /* Allocation request. */
        var hostName string = ""
        var allocId string = ""
        var reclaimAllocId string = ""

        if handle == nil {
                log.Debugf("[%s/%s/%s] EGO client handle is null.", pkg, class, fn)
                return hostName, allocId, reclaimAllocId, nil
        }

        if handle.EGO_handle == nil {
                log.Debugf("[%s/%s/%s] EGO handle is null.", pkg, class, fn)
                return hostName, allocId, reclaimAllocId, nil
        }
       //dimension := map[string]int{"cpu": 4, "memory": 256}
       dimension := map[string]int{"cpu": 2, "memory": 4000}

        aid, rc, error := m.egoModuleClient.GoEGOAllocMD(handle, "xx", consumerId, "Demo", dimension)
        //_, rc, error := ego.GoEGOAllocMD(handle, "xx", "C", "Demo", dimension)


        log.Debugf("[%s/%s/%s] EGO allocation request return code: %v", pkg, class, fn, rc)
        if rc < 0 {
                if error != nil {
                        log.Debugf("[%s/%s/%s] Allocation error: %v", pkg, class, fn, error)
                } else {
                        log.Debugf("[%s/%s/%s] No allocation error provided.", pkg, class, fn)
                }
        } else {
                if error != nil {
                        log.Debugf("[%s/%s/%s] Allocation error: %v", pkg, class, fn, error)
                }
                for {
                        rc, error := m.egoModuleClient.GoEGOSelect(handle)
                        if int(rc) > 0 {
                                break
                        }
                        time.Sleep(1 * time.Second)
                        if error == nil {
                                log.Debugf("[%s/%s/%s] Select complete return code: %v", pkg, class, fn, rc)
                        } else {
                                log.Debugf("[%s/%s/%s] Select complete return code: %v, error: %v", pkg, class, fn, rc, error)
                        }
                }

                result, rc, error := m.egoModuleClient.GoEGORead(handle)
                if rc < 0 {
                        if error == nil {
                                log.Debugf("[%s/%s/%s] Read complete return code: %v", pkg, class, fn, rc)
                        } else {
                                log.Debugf("[%s/%s/%s] Read complete return code: %v, error:", pkg, class, fn, rc, error)
                        }
                }  else {
                        if result.Type == RESOURCE_ADD_MD {
                                //r := result.ReplyBody.(*egomodule.GoAllocReply)
                                var r *egomodule.GoAllocReply = result.ReplyBody.(*egomodule.GoAllocReply)
                                //rb := result.ReplyBody
                                //r := (*egomodule.GoAllocReply) (unsafe.Pointer(&rb))
                                //var strat strategy.PlacementStrategy = s.strategy
                                //pstrat := (*strategy.EGOPlacementStrategy) (unsafe.Pointer(&strat))
                                //err := pstrat.ReleaseAllocation(id)


                                var ndecisions = len(r.Decisions)
                                for i := 0; i < ndecisions; i++ {
                                        hostName = r.Decisions[i].ResourceName
                                        log.Debugf("[%s/%s/%s] Read completed code: %v, placement host name: %s", pkg, class, fn, rc, hostName)
                                        allocId = r.AllocId
                                        break
                                }
                        } else if result.Type == RESOURCE_RECLAIM_MD {
                                var r *egomodule.GoReclaimReply = result.ReplyBody.(*egomodule.GoReclaimReply)
                                var ndecisions = len(r.Decisions)
                                for i := 0; i < ndecisions; i++ {
                                        allocId = aid
                                        hostName = r.Decisions[i].ResourceName
                                        reclaimAllocId = r.Decisions[i].AllocId
                                        log.Debugf("[%s/%s/%s] Read completed code: %v, allocation %v placement on host name: %s reclaiming allocation: %v", pkg, class, fn, rc, allocId, hostName, reclaimAllocId)
                                        break
                                }

                        } else {
                                log.Debugf("[%s/%s/%s] Read completed code: %v.", pkg, class, fn, rc)
                        }
                }
                //TODO: deallocate here
        }
        return hostName, allocId, reclaimAllocId, error
}

func (m *EGOManager) Release(id string) error {
        var fn = "Release"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)
        if handle != nil {
                /* Allocation release. */
                rc1, error1 := m.egoModuleClient.GoEgoReleaseMD(handle, id)
                log.Debugf("[%s/%s/%s] EGO release allocation id: %s return code: %d", pkg, class, fn, id,  rc1)
                if rc1 < 0 {
                    if error1 != nil {
                        log.Debugf("[%s/%s/%s] EGO release allocation id: %s error: %v", pkg, class, fn, id,  error1)
                        return error1
                    }
                }
        } else {
                log.Debugf("[%s/%s/%s] Release ego failed, handle is nil", pkg, class, fn)
        }
        return nil
}
