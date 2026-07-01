//go:build windows

package main

func (s *appState) initV2Features() {
	if s.playCounts == nil {
		s.playCounts = map[string]int{}
	}
	if s.lastPlayed == nil {
		s.lastPlayed = map[string]int64{}
	}
	if s.visualIntensity <= 0 {
		s.visualIntensity = 1.0
	}
	s.installDefaultPack()
	s.loadPresetsFromDisk()
	s.loadPluginPacks()
	if !s.remoteEnabled {
		s.startRemoteControl()
	}
	s.scanLibraryAsync()
}
