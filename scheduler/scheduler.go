package scheduler

import (
        "strings"
        "sync"
        "unsafe"

        "github.com/docker/swarm/cluster"
        "github.com/docker/swarm/scheduler/filter"
        "github.com/docker/swarm/scheduler/node"
        "github.com/docker/swarm/scheduler/strategy"
        log "github.com/Sirupsen/logrus"
)

const (
        pkg   = "scheduler"
        class = "Scheduler"
)

// Scheduler is exported
type Scheduler struct {
        sync.Mutex

        strategy strategy.PlacementStrategy
        filters  []filter.Filter
}

// New is exported
func New(strategy strategy.PlacementStrategy, filters []filter.Filter) *Scheduler {
        return &Scheduler{
                strategy: strategy,
                filters:  filters,
        }
}

// SelectNodeForContainer will find a nice home for our container.
func (s *Scheduler) SelectNodeForContainer(nodes []*node.Node, config *cluster.ContainerConfig) (*node.Node, error) {
        accepted, err := filter.ApplyFilters(s.filters, config, nodes)
        if err != nil {
                return nil, err
        }

        return s.strategy.PlaceContainer(config, accepted)
}

// RemoveAllocation will removed allocation from the scheduler
func (s *Scheduler) RemoveAllocation(id string) error {
        var fn = "RemoveAllocation"
        log.Debugf("[%s/%s/%s] Entering...",  pkg, class, fn)
        var name = s.strategy.Name()
        if name == "ego" {
                var strat strategy.PlacementStrategy = s.strategy
                pstrat := (*strategy.EGOPlacementStrategy) (unsafe.Pointer(&strat))
                err := pstrat.ReleaseAllocation(id)
                if err == nil {
                        log.Debugf("[%s/%s/%s] Removed allocation id %s from scheduler.",  pkg, class, fn, id)
                }
        }
        return nil
}

// Strategy returns the strategy name
func (s *Scheduler) Strategy() string {
        return s.strategy.Name()
}

// Filters returns the list of filter's name
func (s *Scheduler) Filters() string {
        filters := []string{}
        for _, f := range s.filters {
                filters = append(filters, f.Name())
        }

        return strings.Join(filters, ", ")
}

