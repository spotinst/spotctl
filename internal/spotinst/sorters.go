package spotinst

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
