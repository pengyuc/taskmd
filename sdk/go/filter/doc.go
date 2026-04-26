// Package filter applies field-based filters to task collections.
//
// Filters use a "field<op>value" expression syntax with AND logic across
// multiple criteria. The default operator is "=" (exact match). The
// priority and effort fields also support ordering operators: >, >=, <, <=.
//
// Supported fields include status, priority, effort, type, group, tags,
// and assignee.
package filter
