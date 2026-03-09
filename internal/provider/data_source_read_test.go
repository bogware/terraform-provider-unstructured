// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"unstructured": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestDataSourceSourceReadByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sources/src-123":
			_ = json.NewEncoder(w).Encode(SourceConnector{
				ID:        "src-123",
				Name:      "test-source",
				Type:      "s3",
				Config:    map[string]interface{}{"remote_url": "s3://bucket"},
				CreatedAt: "2025-01-01T00:00:00Z",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_source" "test" {
						id = "src-123"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_source.test", "id", "src-123"),
					resource.TestCheckResourceAttr("data.unstructured_source.test", "name", "test-source"),
					resource.TestCheckResourceAttr("data.unstructured_source.test", "type", "s3"),
					resource.TestCheckResourceAttr("data.unstructured_source.test", "created_at", "2025-01-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestDataSourceSourceReadByName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sources/":
			_ = json.NewEncoder(w).Encode([]SourceConnector{
				{
					ID:        "src-456",
					Name:      "my-source",
					Type:      "azure",
					Config:    map[string]interface{}{"remote_url": "az://container"},
					CreatedAt: "2025-02-01T00:00:00Z",
				},
				{
					ID:        "src-789",
					Name:      "other-source",
					Type:      "gcs",
					Config:    map[string]interface{}{},
					CreatedAt: "2025-03-01T00:00:00Z",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_source" "test" {
						name = "my-source"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_source.test", "id", "src-456"),
					resource.TestCheckResourceAttr("data.unstructured_source.test", "name", "my-source"),
					resource.TestCheckResourceAttr("data.unstructured_source.test", "type", "azure"),
				),
			},
		},
	})
}

func TestDataSourceSourceNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_source" "test" {
						id = "nonexistent"
					}
				`, ts.URL),
				ExpectError: regexp.MustCompile(`Not Found`),
			},
		},
	})
}

func TestDataSourceDestinationReadByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/destinations/dst-123":
			_ = json.NewEncoder(w).Encode(DestinationConnector{
				ID:        "dst-123",
				Name:      "test-destination",
				Type:      "pinecone",
				Config:    map[string]interface{}{"api_key": "pk-xxx"},
				CreatedAt: "2025-01-01T00:00:00Z",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_destination" "test" {
						id = "dst-123"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "id", "dst-123"),
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "name", "test-destination"),
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "type", "pinecone"),
				),
			},
		},
	})
}

func TestDataSourceDestinationReadByName(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/destinations/":
			_ = json.NewEncoder(w).Encode([]DestinationConnector{
				{
					ID:        "dst-456",
					Name:      "my-dest",
					Type:      "s3",
					Config:    map[string]interface{}{"remote_url": "s3://output"},
					CreatedAt: "2025-02-01T00:00:00Z",
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_destination" "test" {
						name = "my-dest"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "id", "dst-456"),
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "name", "my-dest"),
					resource.TestCheckResourceAttr("data.unstructured_destination.test", "type", "s3"),
				),
			},
		},
	})
}

func TestDataSourceWorkflowRead(t *testing.T) {
	updatedAt := "2025-06-01T00:00:00Z"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/workflows/wf-123":
			_ = json.NewEncoder(w).Encode(Workflow{
				ID:            "wf-123",
				Name:          "test-workflow",
				Sources:       []string{"src-1"},
				Destinations:  []string{"dst-1"},
				WorkflowType:  "custom",
				WorkflowNodes: []WorkflowNode{{Name: "partitioner", Type: "partition", Subtype: "vlm"}},
				Schedule: &WorkflowSchedule{
					CrontabEntries: []CrontabEntry{{CronExpression: "0 0 * * *"}},
				},
				Status:       "active",
				CreatedAt:    "2025-01-01T00:00:00Z",
				UpdatedAt:    &updatedAt,
				ReprocessAll: true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_workflow" "test" {
						id = "wf-123"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "id", "wf-123"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "name", "test-workflow"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "source_id", "src-1"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "destination_id", "dst-1"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "workflow_type", "custom"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "status", "active"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "schedule", "0 0 * * *"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "reprocess_all", "true"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "created_at", "2025-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.unstructured_workflow.test", "updated_at", "2025-06-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestDataSourceJobRead(t *testing.T) {
	runtime := "45s"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/jobs/job-123":
			_ = json.NewEncoder(w).Encode(Job{
				ID:           "job-123",
				WorkflowID:   "wf-456",
				WorkflowName: "my-workflow",
				Status:       "COMPLETED",
				CreatedAt:    "2025-01-15T10:00:00Z",
				Runtime:      &runtime,
				JobType:      "scheduled",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_job" "test" {
						id = "job-123"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_job.test", "id", "job-123"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "workflow_id", "wf-456"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "workflow_name", "my-workflow"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "status", "COMPLETED"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "created_at", "2025-01-15T10:00:00Z"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "runtime", "45s"),
					resource.TestCheckResourceAttr("data.unstructured_job.test", "job_type", "scheduled"),
				),
			},
		},
	})
}

func TestDataSourceTemplateRead(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/templates/tmpl-123":
			_ = json.NewEncoder(w).Encode(Template{
				ID:           "tmpl-123",
				Name:         "Basic Pipeline",
				Description:  "A basic document processing pipeline",
				WorkflowType: "platinum",
				WorkflowNodes: []WorkflowNode{
					{Name: "partitioner", Type: "partition", Subtype: "auto"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					provider "unstructured" {
						api_key = "test-key"
						api_url = %q
					}
					data "unstructured_template" "test" {
						id = "tmpl-123"
					}
				`, ts.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.unstructured_template.test", "id", "tmpl-123"),
					resource.TestCheckResourceAttr("data.unstructured_template.test", "name", "Basic Pipeline"),
					resource.TestCheckResourceAttr("data.unstructured_template.test", "description", "A basic document processing pipeline"),
					resource.TestCheckResourceAttr("data.unstructured_template.test", "workflow_type", "platinum"),
				),
			},
		},
	})
}
