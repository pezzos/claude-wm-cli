// Package application provides filter adapters that implement model.Filter interface.
// These adapters wrap entity filters to make them compatible with the model.Filter interface.
package application

import (
	"fmt"
	"strings"
	"time"

	"claude-wm-cli/internal/entity"
	"claude-wm-cli/internal/model"
)

// statusFilterAdapter wraps entity.StatusFilter to implement model.Filter
type statusFilterAdapter struct {
	inner *entity.StatusFilter
}

// Apply returns true if the entity matches the filter criteria
func (s *statusFilterAdapter) Apply(entity interface{}) bool {
	return s.inner.Apply(entity)
}

// ToSQL converts the filter to SQL WHERE clause
func (s *statusFilterAdapter) ToSQL() (string, []interface{}) {
	return "status = ?", []interface{}{string(s.inner.Status)}
}

// Validate checks if the filter criteria are valid
func (s *statusFilterAdapter) Validate() error {
	if !s.inner.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", s.inner.Status)
	}
	return nil
}

// priorityFilterAdapter wraps entity.PriorityFilter to implement model.Filter
type priorityFilterAdapter struct {
	inner *entity.PriorityFilter
}

// Apply returns true if the entity matches the filter criteria
func (p *priorityFilterAdapter) Apply(entity interface{}) bool {
	return p.inner.Apply(entity)
}

// ToSQL converts the filter to SQL WHERE clause
func (p *priorityFilterAdapter) ToSQL() (string, []interface{}) {
	return "priority = ?", []interface{}{string(p.inner.Priority)}
}

// Validate checks if the filter criteria are valid
func (p *priorityFilterAdapter) Validate() error {
	if !p.inner.Priority.IsValid() {
		return fmt.Errorf("invalid priority: %s", p.inner.Priority)
	}
	return nil
}

// dateRangeFilterAdapter wraps entity.DateRangeFilter to implement model.Filter
type dateRangeFilterAdapter struct {
	inner *entity.DateRangeFilter
}

// Apply returns true if the entity matches the filter criteria
func (d *dateRangeFilterAdapter) Apply(entity interface{}) bool {
	return d.inner.Apply(entity)
}

// ToSQL converts the filter to SQL WHERE clause
func (d *dateRangeFilterAdapter) ToSQL() (string, []interface{}) {
	var conditions []string
	var args []interface{}
	
	if d.inner.After != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *d.inner.After)
	}
	if d.inner.Before != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *d.inner.Before)
	}
	
	if len(conditions) == 0 {
		return "1=1", []interface{}{}
	}
	
	return strings.Join(conditions, " AND "), args
}

// Validate checks if the filter criteria are valid
func (d *dateRangeFilterAdapter) Validate() error {
	if d.inner.After != nil && d.inner.Before != nil && d.inner.After.After(*d.inner.Before) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

// multiFilterAdapter wraps entity.MultiFilter to implement model.Filter
type multiFilterAdapter struct {
	inner *entity.MultiFilter
}

// Apply returns true if the entity matches the filter criteria
func (m *multiFilterAdapter) Apply(entity interface{}) bool {
	return m.inner.Apply(entity)
}

// ToSQL converts the filter to SQL WHERE clause
func (m *multiFilterAdapter) ToSQL() (string, []interface{}) {
	if len(m.inner.Filters) == 0 {
		return "1=1", []interface{}{}
	}
	
	var conditions []string
	var args []interface{}
	
	for _, filter := range m.inner.Filters {
		// All filters in MultiFilter.Filters should already be model.Filter
		var f model.Filter
		if modelFilter, ok := filter.(model.Filter); ok {
			f = modelFilter
		} else {
			continue // Skip unsupported filter types
		}
		
		condition, filterArgs := f.ToSQL()
		if condition != "" && condition != "1=1" {
			conditions = append(conditions, fmt.Sprintf("(%s)", condition))
			args = append(args, filterArgs...)
		}
	}
	
	if len(conditions) == 0 {
		return "1=1", []interface{}{}
	}
	
	operator := "AND"  // MultiFilter from entity uses AND logic by default
	
	return strings.Join(conditions, fmt.Sprintf(" %s ", operator)), args
}

// Validate checks if the filter criteria are valid
func (m *multiFilterAdapter) Validate() error {
	for _, filter := range m.inner.Filters {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("nested filter validation failed: %w", err)
		}
	}
	return nil
}

// Helper functions to create filter adapters

func newStatusFilterAdapter(status model.Status) model.Filter {
	return &statusFilterAdapter{
		inner: &entity.StatusFilter{Status: status},
	}
}

func newPriorityFilterAdapter(priority model.Priority) model.Filter {
	return &priorityFilterAdapter{
		inner: &entity.PriorityFilter{Priority: priority},
	}
}

func newDateRangeFilterAdapter(startDate, endDate *time.Time) model.Filter {
	return &dateRangeFilterAdapter{
		inner: &entity.DateRangeFilter{
			After:  startDate,
			Before: endDate,
		},
	}
}

func newMultiFilterAdapter(filters []interface{}, operator string) model.Filter {
	// Convert filters to model.Filter slice
	modelFilters := make([]model.Filter, 0, len(filters))
	for _, f := range filters {
		if mf, ok := f.(model.Filter); ok {
			modelFilters = append(modelFilters, mf)
		}
	}
	
	return &multiFilterAdapter{
		inner: &entity.MultiFilter{
			Filters: modelFilters,
		},
	}
}