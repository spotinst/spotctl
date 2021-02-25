package spot

type OceanClustersSorter struct {
	Clusters []*OceanCluster
}

func (s *OceanClustersSorter) Len() int {
	return len(s.Clusters)
}

func (s *OceanClustersSorter) Swap(i, j int) {
	s.Clusters[i], s.Clusters[j] = s.Clusters[j], s.Clusters[i]
}

func (s *OceanClustersSorter) Less(i, j int) bool {
	return s.Clusters[i].UpdatedAt.After(s.Clusters[j].UpdatedAt)
}

type OceanLaunchSpecsSorter struct {
	LaunchSpecs []*OceanLaunchSpec
}

func (s *OceanLaunchSpecsSorter) Len() int {
	return len(s.LaunchSpecs)
}

func (s *OceanLaunchSpecsSorter) Swap(i, j int) {
	s.LaunchSpecs[i], s.LaunchSpecs[j] = s.LaunchSpecs[j], s.LaunchSpecs[i]
}

func (s *OceanLaunchSpecsSorter) Less(i, j int) bool {
	return s.LaunchSpecs[i].UpdatedAt.After(s.LaunchSpecs[j].UpdatedAt)
}

type OceanRolloutsSorter struct {
	Rollouts []*OceanRollout
}

func (s *OceanRolloutsSorter) Len() int {
	return len(s.Rollouts)
}

func (s *OceanRolloutsSorter) Swap(i, j int) {
	s.Rollouts[i], s.Rollouts[j] = s.Rollouts[j], s.Rollouts[i]
}

func (s *OceanRolloutsSorter) Less(i, j int) bool {
	return s.Rollouts[i].UpdatedAt.After(s.Rollouts[j].UpdatedAt)
}

type WaveClustersSorter struct {
	Clusters []*WaveCluster
}

func (s *WaveClustersSorter) Len() int {
	return len(s.Clusters)
}

func (s *WaveClustersSorter) Swap(i, j int) {
	s.Clusters[i], s.Clusters[j] = s.Clusters[j], s.Clusters[i]
}

func (s *WaveClustersSorter) Less(i, j int) bool {
	return s.Clusters[i].UpdatedAt.After(s.Clusters[j].UpdatedAt)
}
