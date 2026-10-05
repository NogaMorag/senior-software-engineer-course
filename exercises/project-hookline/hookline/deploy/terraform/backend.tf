# The local backend, written out so that the state location is a decision you can see. The state can
# hold the database URL (a sensitive variable): it is git-ignored, and a team would use a remote backend
# with locking and encryption instead (chapter 22, section 22.5).
terraform {
  backend "local" {
    path = "state/terraform.tfstate"
  }
}
