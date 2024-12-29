terraform {
  backend "s3" {
    bucket  = "bitkitchen-tf-state"
    key     = "myxcaliapp/session-test"
    region  = "eu-west-1"
    encrypt = true
  }
}
