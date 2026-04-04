package usecase

import "testing"

func TestInitUseCase_Execute_CreatesDefaultConfig(t *testing.T) {
	uc := NewInitUseCase()

	config, err := uc.Execute("/home/user/.distill", "/home/user/.distill-trash")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if config.Core.Version != "1" {
		t.Errorf("config version = %q, want %q", config.Core.Version, "1")
	}
	if config.Core.ObjectsFormat != "plain" {
		t.Errorf("objects format = %q, want %q", config.Core.ObjectsFormat, "plain")
	}
	if config.Store.Home != "/home/user/.distill" {
		t.Errorf("store home = %q, want %q", config.Store.Home, "/home/user/.distill")
	}
	if config.Store.TrashPath != "/home/user/.distill-trash" {
		t.Errorf("trash path = %q, want %q", config.Store.TrashPath, "/home/user/.distill-trash")
	}
	if config.Checkout.Overwrite != "ask" {
		t.Errorf("overwrite = %q, want %q", config.Checkout.Overwrite, "ask")
	}
	if config.Log.Format != "text" {
		t.Errorf("log format = %q, want %q", config.Log.Format, "text")
	}
	if config.Log.Level != "info" {
		t.Errorf("log level = %q, want %q", config.Log.Level, "info")
	}
	if config.Normalize.CRLFToLF != true {
		t.Error("crlf_to_lf should default to true")
	}
}

func TestInitUseCase_Execute_AlreadyInitialized(t *testing.T) {
	uc := NewInitUseCase()

	_, err := uc.Execute("/home/user/.distill", "/home/user/.distill-trash")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	_, err = uc.Execute("/home/user/.distill", "/home/user/.distill-trash")
	if err == nil {
		t.Error("second init should return error (already initialized)")
	}
}
