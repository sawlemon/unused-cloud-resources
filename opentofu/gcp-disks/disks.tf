
resource "google_compute_disk" "us-c1a-disks" {
  count = 10
  name  = "pd-us-c1a-${count.index}"
  zone  = "us-central1-a"
  size  = 4

  labels = {
    environment = "dev"
  }
}

resource "google_compute_disk" "us-c1b-disks" {
  count = 10
  name  = "pd-us-c1b-${count.index}"
  size  = 4
  zone  = "us-central1-b"

  labels = {
    environment = "dev"
  }
}