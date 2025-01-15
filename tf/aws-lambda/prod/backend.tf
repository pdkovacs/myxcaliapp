terraform {
  backend "s3" {
    bucket  = "bitkitchen-tf-state"
    key     = "myxcaliapp/prod"
    region  = "eu-west-1"
    encrypt = true
  }
}
