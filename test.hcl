group_name = "test"

component_from_image "errordeveloper/test:latest" {
  flavor = "minimal"
}

component_from_image "errordeveloper/test:latest" {
  name = "boo"
}
