# The local backend is the default; writing it out makes the state location a
# decision you can see. A team would use a remote backend with locking instead
# (section 22.5). The state file is git-ignored: it can contain secrets.
terraform {
  backend "local" {
    path = "state/terraform.tfstate"
  }
}
