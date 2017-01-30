group_name = "test"

component_from_image "errordeveloper/test:latest" {
  flavor = "minimal"
  port = 443
}

component_from_image "errordeveloper/test:latest" {
  name = "boo"
}

component_template "foo" {
  image = "errordeveloper/foo"
  name = "foofoo"
  port = 123
}

component_from_template "foo" {
  port = 78
}

component_from_template "foo" {
  name = "barbar"
}

component_from_template "foo" {
  image = "errordeveloper/foo-plus"
}
