
resource google_storage_bucket "main-bucket" {
  project            = "nicolabs-gcp-assignment"
  name               = "main-bucket"
  location           = "EU"
  storage_class      = "MULTI_REGIONAL"
  bucket_policy_only = false
  force_destroy      = false

  versioning {
    enabled = false
  }
}

resource google_storage_bucket "second-bucket" {
  project            = "nicolabs-gcp-assignment"
  name               = "second-bucket"
  location           = "EU"
  storage_class      = "MULTI_REGIONAL"
  bucket_policy_only = false
  force_destroy      = false

  versioning {
    enabled = false
  }
}

resource "google_service_account" "service_account" {
  project      = "nicolabs-gcp-assignment"
  account_id   = "project-account"
  display_name = "Service Account"
}

resource "google_project_iam_member" "storage_events_cloudrun" {
  for_each = toset([
    "roles/run.invoker",
    "roles/storage.objectAdmin",
    "roles/pubsub.publisher",
    "roles/logging.logWriter",
    "roles/iam.serviceAccountTokenCreator"
  ])

  project = "nicolabs-gcp-assignment"
  member  = "serviceAccount:${google_service_account.service_account.email}"
  role    = each.value
}


resource "google_pubsub_topic" "storage_events" {
  project = "nicolabs-gcp-assignment"
  name    = "storage-events"
}

resource "google_pubsub_subscription" "push" {
  name    = "push-subscription"
  project = "nicolabs-gcp-assignment"
  topic   = google_pubsub_topic.storage_events.name

  ack_deadline_seconds = 20

  push_config {
    push_endpoint = google_cloud_run_service.run.status[0].url
    oidc_token {
      service_account_email = google_service_account.service_account.email
    }
  }
}

# A push subscription pushes bucket events to this service,
# which forwards them to tasks
resource "google_cloud_run_service" "run" {
  provider = google-beta
  name     = "storage-events"
  project  = "nicolabs-gcp-assignment"
  location = "europe-west1"

  metadata {
    namespace = "nicolabs-gcp-assignment"
  }

  template {
    spec {
      service_account_name = google_service_account.service_account.email
#containers {
#image = "gcr.io"nicolabs-gcp-assignment"storage-events"
#}
    }
  }
}

// Bucket triggers
data "google_storage_project_service_account" "gcs_account" {
  project = "nicolabs-gcp-assignment"
}

resource "google_storage_notification" "notification" {
  bucket         = "main-bucket"
  payload_format = "JSON_API_V1"
  topic          = google_pubsub_topic.storage_events.id
  event_types    = ["OBJECT_FINALIZE", "OBJECT_DELETE"]
  depends_on     = [google_pubsub_topic_iam_binding.binding]
}

resource "google_pubsub_topic_iam_binding" "binding" {
  project = "nicolabs-gcp-assignment"
  topic   = google_pubsub_topic.storage_events.id
  role    = "roles/pubsub.publisher"
  members = ["serviceAccount:${data.google_storage_project_service_account.gcs_account.email_address}"]
}
