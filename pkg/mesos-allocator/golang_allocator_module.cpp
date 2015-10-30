#include <mesos/master/allocator.hpp>
#include <mesos/module/allocator.hpp>
#include <stout/try.hpp>

using mesos::master::allocator::Allocator;

namespace mesos {
namespace golang {


static Allocator* createGolangAllocator(const Parameters& parameters)
  {
    return NULL;
    // Try<Allocator*> allocator = GolangAllocator::create();
    // if (allocator.isError()) {
    //   return NULL;
    // }

    // return allocator.get();
  };

} // namespace golang {
} // namespace mesos {

mesos::modules::Module<Allocator> github_com_mdevilliers_golang_allocator(
    MESOS_MODULE_API_VERSION,
    MESOS_VERSION,
    "Mesos Contributor",
    "engineer@example.com",
    "External Allocator module.",
    NULL,
    mesos::golang::createGolangAllocator);