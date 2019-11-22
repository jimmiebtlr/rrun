locals {
  cluster_project = "rdev-257506"
}

provider "google" {}

resource "google_container_cluster" "primary" {
  project  = local.cluster_project
  name     = "developer-resources"
  location = "us-central1"

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1
}

resource "google_container_node_pool" "primary_preemptible_nodes" {
  project    = local.cluster_project
  name       = "dev-resources"
  location   = "us-central1"
  cluster    = google_container_cluster.primary.name

  autoscaling {
    min_node_count = 0
    max_node_count = 1
  }

  node_config {
    preemptible  = true
    machine_type = "n1-standard-1"

    metadata = {
      disable-legacy-endpoints = "true"
    }

    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }
}
