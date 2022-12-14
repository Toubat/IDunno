package ring

import "mp4/api"

/* MembershipList
 * - List of processes in the ring
 * - Implements sort.Interface
 * - Sorted by timestamp
 */
type MembershipList []*api.Process

func (m MembershipList) Len() int {
	return len(m)
}

func (m MembershipList) Less(i, j int) bool {
	return m[i].JoinTime.AsTime().Before(m[j].JoinTime.AsTime())
}

func (m MembershipList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m MembershipList) Filter(predicate func(*api.Process) bool) MembershipList {
	filtered := make(MembershipList, 0)

	for _, process := range m {
		if predicate(process) {
			filtered = append(filtered, process)
		}
	}

	return filtered
}
