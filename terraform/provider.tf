provider "google-beta" {
  credentials = "${file("service-account.json")}"
  project = "nicolabs-gcp-assignment"
  region  = "europe-west1"
  zone    = "europe-west1-b"
  version = "~> 3.5"
}