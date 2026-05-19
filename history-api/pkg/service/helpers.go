package service

import (
	"fmt"
	"strings"

	"github.com/runtime-radar/runtime-radar/history-api/api"
	"github.com/runtime-radar/runtime-radar/history-api/pkg/database/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultStatsOrder = "severity desc, count desc"

// makeRuntimeEventFilter prepares values and patterns from rf to be used in SQL query and constructs gorm expression.
// nolint:goconst
func makeRuntimeEventFilter(rf *api.RuntimeFilter) (clause.Expr, error) {
	var (
		eventTypeEq         = make([]string, 0, len(rf.GetEventType()))
		kprobeFuncLike      = make([]string, 0, len(rf.GetKprobeFunctionName()))
		podNamespaceLike    = make([]string, 0, len(rf.GetProcessPodNamespace()))
		podNameLike         = make([]string, 0, len(rf.GetProcessPodName()))
		nodeNameLike        = make([]string, 0, len(rf.GetNodeName()))
		containerNameLike   = make([]string, 0, len(rf.GetProcessPodContainerName()))
		imageNameLike       = make([]string, 0, len(rf.GetProcessPodContainerImageName()))
		processBinaryLike   = make([]string, 0, len(rf.GetProcessBinary()))
		processArgsLike     = make([]string, 0, len(rf.GetProcessArguments()))
		threatsDetectorsHas = make([]string, 0, len(rf.GetThreatsDetectors()))
		rulesHas            = make([]string, 0, len(rf.GetRules()))
	)

	argsNum := len(rf.GetEventType()) +
		len(rf.GetKprobeFunctionName()) +
		len(rf.GetProcessPodNamespace()) +
		len(rf.GetProcessPodName()) +
		len(rf.GetNodeName()) +
		len(rf.GetProcessPodContainerName()) +
		len(rf.GetProcessPodContainerImageName()) +
		len(rf.GetProcessBinary()) +
		len(rf.GetProcessArguments()) +
		len(rf.GetThreatsDetectors()) +
		len(rf.GetRules())*2 // we're searching in block_by and notify_by at the same time so x2 args are needed

	if rf.GetProcessExecId() != "" {
		argsNum++
	}
	if rf.GetProcessParentExecId() != "" {
		argsNum++
	}

	r1 := strings.NewReplacer(
		"%", "", // avoid malicious requests
		"_", `\_`, // escape special symbol _
	)
	r2 := strings.NewReplacer(
		"**", "%",
		"*", "%",
		"?", "_",
	)

	// since we want to keep replacers' internal states, we initialize them once and keep this function local
	prepareTemplate := func(s string) string {
		tpl := r1.Replace(s)
		tpl = r2.Replace(tpl)
		return tpl
	}

	args := make([]interface{}, 0, argsNum)
	for _, t := range rf.GetEventType() {
		eventTypeEq = append(eventTypeEq, "event_type = ?")
		args = append(args, t)
	}

	for _, fn := range rf.GetKprobeFunctionName() {
		kprobeFuncLike = append(kprobeFuncLike, "kprobe_function_name LIKE ?")
		args = append(args, prepareTemplate(fn))
	}

	for _, pns := range rf.GetProcessPodNamespace() {
		podNamespaceLike = append(podNamespaceLike, "process_pod_namespace LIKE ?")
		args = append(args, prepareTemplate(pns))
	}

	for _, pn := range rf.GetProcessPodName() {
		podNameLike = append(podNameLike, "process_pod_name LIKE ?")
		args = append(args, prepareTemplate(pn))
	}

	for _, nn := range rf.GetNodeName() {
		nodeNameLike = append(nodeNameLike, "node_name LIKE ?")
		args = append(args, prepareTemplate(nn))
	}

	for _, cn := range rf.GetProcessPodContainerName() {
		containerNameLike = append(containerNameLike, "process_pod_container_name LIKE ?")
		args = append(args, prepareTemplate(cn))
	}

	for _, in := range rf.GetProcessPodContainerImageName() {
		imageNameLike = append(imageNameLike, "process_pod_container_image_name LIKE ?")
		args = append(args, prepareTemplate(in))
	}

	for _, b := range rf.GetProcessBinary() {
		processBinaryLike = append(processBinaryLike, "process_binary LIKE ?")
		args = append(args, prepareTemplate(b))
	}

	for _, a := range rf.GetProcessArguments() {
		processArgsLike = append(processArgsLike, "process_arguments LIKE ?")
		args = append(args, prepareTemplate(a))
	}

	for _, td := range rf.GetThreatsDetectors() {
		threatsDetectorsHas = append(threatsDetectorsHas, "has(threats_detectors, ?)")
		args = append(args, td)
	}

	for _, r := range rf.GetRules() {
		rulesHas = append(rulesHas, "has(block_by, ?) OR has(notify_by, ?)")
		args = append(args, r, r)
	}

	sql := makeAndWhereClause(
		eventTypeEq,
		kprobeFuncLike,
		podNamespaceLike,
		podNameLike,
		nodeNameLike,
		containerNameLike,
		imageNameLike,
		processBinaryLike,
		processArgsLike,
		threatsDetectorsHas,
		rulesHas,
	)

	if from := rf.GetPeriod().GetFrom(); from != nil {
		if sql != "" {
			sql += " AND "
		}

		asTime := from.AsTime()

		// toDateTime64 has to be used because DateTime64 cannot be automatically converted from string. See https://clickhouse.com/docs/en/sql-reference/data-types/datetime64 for details.
		sql += "registered_at > toDateTime64(?, 9, ?)"
		args = append(args, asTime.Format(clickhouse.DateTimeFormat), asTime.Location().String())
	}

	if to := rf.GetPeriod().GetTo(); to != nil {
		if sql != "" {
			sql += " AND "
		}

		asTime := to.AsTime()

		// toDateTime64 has to be used because DateTime64 cannot be automatically converted from string. See https://clickhouse.com/docs/en/sql-reference/data-types/datetime64 for details.
		sql += "registered_at < toDateTime64(?, 9, ?)"
		args = append(args, asTime.Format(clickhouse.DateTimeFormat), asTime.Location().String())
	}

	if rf.HasThreats != nil {
		if sql != "" {
			sql += " AND "
		}

		if *rf.HasThreats {
			sql += "threats IS NOT NULL"
		} else {
			sql += "threats IS NULL"
		}
	}

	if execID := rf.GetProcessExecId(); execID != "" {
		if sql != "" {
			sql += " AND "
		}

		sql += "process_exec_id = ?"
		args = append(args, execID)
	}

	if parentExecID := rf.GetProcessParentExecId(); parentExecID != "" {
		if sql != "" {
			sql += " AND "
		}

		sql += "process_parent_exec_id = ?"
		args = append(args, parentExecID)
	}

	if rf.HasIncident != nil {
		if sql != "" {
			sql += " AND "
		}

		if *rf.HasIncident {
			sql += "is_incident"
		} else {
			sql += "NOT is_incident"
		}
	}

	return gorm.Expr(sql, args...), nil
}

func makeDetectorRatingFilter(rf *api.DetectorRatingReq_Filter) clause.Expr {
	if rf == nil {
		return clause.Expr{}
	}

	var names, namespaces, containers []string
	var args []any

	if t := rf.GetPodName(); t != "" {
		names = append(names, "process_pod_name = ?")
		args = append(args, t)
	}

	if t := rf.GetPodNamespace(); t != "" {
		namespaces = append(namespaces, "process_pod_namespace = ?")
		args = append(args, t)
	}

	if t := rf.GetContainerName(); t != "" {
		containers = append(containers, "process_pod_container_name = ?")
		args = append(args, t)
	}

	return gorm.Expr(makeAndWhereClause(names, namespaces, containers), args...)
}

// makeAndWhereClause returns SQL string with multiple conditions joint via AND.
// Every subClauses' element must represent slice of conditions which will be joint via OR.
// For example, makeAndWhereClause([]string{"a = b", "c = d"}, []string{"e = f"}) will produce the following expression:
// (a = b OR c = d) AND (e = f)
func makeAndWhereClause(subClauses ...[]string) string {
	parts := make([]string, 0, len(subClauses))

	for _, c := range subClauses {
		if len(c) != 0 {
			part := "(" + strings.Join(c, " OR ") + ")"
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, " AND ")
}

func makeOrder(sorts []*api.Sort) string {
	if len(sorts) == 0 {
		return ""
	}

	sb := &strings.Builder{}

	for i, s := range sorts {
		sb.WriteString(s.Field)
		sb.WriteRune(' ')
		sb.WriteString(s.Key)

		if i < len(sorts)-1 {
			sb.WriteRune(',')
		}
	}

	return sb.String()
}

func makeOrderSlice(sorts []*api.Sort) []string {
	if len(sorts) == 0 {
		return nil
	}

	sb := make([]string, 0, countSortsSize(sorts))

	for _, s := range sorts {
		if len(s.Field) > 0 && len(s.Key) > 0 {
			sb = append(sb, fmt.Sprintf("%s %s", s.Field, s.Key))
		}
	}

	return sb
}

func countSortsSize(sorts []*api.Sort) int {
	count := 0
	for _, s := range sorts {
		if len(s.Field) > 0 && len(s.Key) > 0 {
			count++
		}
	}
	return count
}
