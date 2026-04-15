package service

import (
	"reflect"
	"testing"

	"api_zhuanfa/internal/model"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newAPIKeyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.ApiKey{}); err != nil {
		t.Fatalf("migrate api keys: %v", err)
	}
	return db
}

func TestParseAllowedUpstreamIDs(t *testing.T) {
	got := ParseAllowedUpstreamIDs(" 2,1,2,abc,0,3 ")
	want := []uint{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected ids: got %v want %v", got, want)
	}
}

func TestCheckUpstreamAllowed(t *testing.T) {
	svc := &ApiKeyService{}

	if !svc.CheckUpstreamAllowed(&model.ApiKey{AllowedUpstreams: ""}, 99) {
		t.Fatal("empty allowed_upstreams should allow all upstreams")
	}
	if !svc.CheckUpstreamAllowed(&model.ApiKey{AllowedUpstreams: " 2,1,2 "}, 1) {
		t.Fatal("expected upstream 1 to be allowed")
	}
	if svc.CheckUpstreamAllowed(&model.ApiKey{AllowedUpstreams: " 2,1,2 "}, 3) {
		t.Fatal("expected upstream 3 to be denied")
	}
}

func TestGenerateAndUpdateNormalizeAllowedUpstreams(t *testing.T) {
	db := newAPIKeyTestDB(t)
	svc := NewApiKeyService(db)

	_, item, err := svc.Generate(1, "demo", 0, nil, " 2,1,2,abc,3 ")
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}
	if item.AllowedUpstreams != "1,2,3" {
		t.Fatalf("unexpected normalized upstreams after generate: %q", item.AllowedUpstreams)
	}

	var stored model.ApiKey
	if err := db.First(&stored, item.ID).Error; err != nil {
		t.Fatalf("load stored api key: %v", err)
	}
	if stored.AllowedUpstreams != "1,2,3" {
		t.Fatalf("unexpected stored upstreams after generate: %q", stored.AllowedUpstreams)
	}

	if err := svc.Update(item.ID, map[string]interface{}{"allowed_upstreams": "4, 3,4,abc"}); err != nil {
		t.Fatalf("update allowed_upstreams: %v", err)
	}
	if err := db.First(&stored, item.ID).Error; err != nil {
		t.Fatalf("reload api key after string update: %v", err)
	}
	if stored.AllowedUpstreams != "3,4" {
		t.Fatalf("unexpected stored upstreams after string update: %q", stored.AllowedUpstreams)
	}

	if err := svc.Update(item.ID, map[string]interface{}{"allowed_upstream_ids": []interface{}{6.0, "5", 6.0}}); err != nil {
		t.Fatalf("update allowed_upstream_ids: %v", err)
	}
	if err := db.First(&stored, item.ID).Error; err != nil {
		t.Fatalf("reload api key after array update: %v", err)
	}
	if stored.AllowedUpstreams != "5,6" {
		t.Fatalf("unexpected stored upstreams after array update: %q", stored.AllowedUpstreams)
	}
}

func TestUpdateRejectsInvalidAllowedUpstreamIDs(t *testing.T) {
	db := newAPIKeyTestDB(t)
	svc := NewApiKeyService(db)

	_, item, err := svc.Generate(1, "demo", 0, nil, "")
	if err != nil {
		t.Fatalf("generate api key: %v", err)
	}

	if err := svc.Update(item.ID, map[string]interface{}{"allowed_upstream_ids": []interface{}{"bad"}}); err == nil {
		t.Fatal("expected invalid allowed_upstream_ids to return error")
	}
}
