# Variables like in terraform. Could do module outputs as well, for service
# ports, etc, if needed?
variable "env_name" {
  default = "prod"
}

variable "zipkin" {
  default = "http://zipkin:9411/api/v1/spans"
}

# Using nesting to denote belonging to a namespace, could also put it on each
# individual resource, but that seems onerous. Downside of the nesting is you
# have to check you haven't done: `namespace "foo" { namespace "bar" {} }`
namespace "sock-shop" {
  module "cart" {
    source = "../modules/cart"
    zipkin = "${var.zipkin}"
  }

  module "catalogue" {
    source = "../modules/catalogue"

    zipkin = "${var.zipkin}"
    mysqlpass = "fake_password"
    mysqldb = "socksdb"
  }

  # ... More modules for the different services here.
  #
  # Maybe modules should work more like nginx-config-imports where it just
  # blats the source in. But, it would be nice to be able to import
  # "functions", then have them be parameterizable, which is sort of what the
  # module syntax gives you...
}
