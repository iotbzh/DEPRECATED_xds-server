package xdsserver

import "github.com/iotbzh/xds-server/lib/xsapiv1"

type FolderEventCBData map[string]interface{}
type FolderEventCB func(cfg *xsapiv1.FolderConfig, data *FolderEventCBData)

// IFOLDER Folder interface
type IFOLDER interface {
	NewUID(suffix string) string                                          // Get a new folder UUID
	Add(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error)              // Add a new folder
	GetConfig() xsapiv1.FolderConfig                                        // Get folder public configuration
	GetFullPath(dir string) string                                        // Get folder full path
	ConvPathCli2Svr(s string) string                                      // Convert path from Client to Server
	ConvPathSvr2Cli(s string) string                                      // Convert path from Server to Client
	Remove() error                                                        // Remove a folder
	Update(cfg xsapiv1.FolderConfig) (*xsapiv1.FolderConfig, error)           // Update a new folder
	RegisterEventChange(cb *FolderEventCB, data *FolderEventCBData) error // Request events registration (sent through WS)
	UnRegisterEventChange() error                                         // Un-register events
	Sync() error                                                          // Force folder files synchronization
	IsInSync() (bool, error)                                              // Check if folder files are in-sync
}
