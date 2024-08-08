terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.40.0"
    }
  }
}

provider "google" {
  project = "finops-accelerator"
  region  = "us-central1"
  // Make sure to authenticate and point to the ADC credentials file below
  credentials = "/Users/sala/.config/gcloud/application_default_credentials.json"
}