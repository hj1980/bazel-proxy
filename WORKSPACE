load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "d6b2513456fe2229811da7eb67a444be7785f5323c6708b38d851d2b51e54d83",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.30.0/rules_go-v0.30.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.30.0/rules_go-v0.30.0.zip",
    ],
)
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
go_rules_dependencies()
go_register_toolchains(version = "host")

http_archive(
    name = "bazel_gazelle",
    sha256 = "de69a09dc70417580aabf20a28619bb3ef60d038470c7cf8442fafcf627c21cb",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.24.0/bazel-gazelle-v0.24.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.24.0/bazel-gazelle-v0.24.0.tar.gz",
    ],
)
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
gazelle_dependencies()

http_archive(
    name = "com_google_protobuf",
    sha256 = "3bd7828aa5af4b13b99c191e8b1e884ebfa9ad371b0ce264605d347f135d2568",
    strip_prefix = "protobuf-3.19.4",
    urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.19.4.tar.gz"],
)
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
protobuf_deps()

http_archive(
    name = "bazel_remote_apis",
    sha256 = "e9c75f44ef65a91a254113c2ab2ea49d959bea30244e84a6e641c61094e2f1b2",
    strip_prefix = "remote-apis-04784f4a830cc0df1f419a492cde9fc323f728db",
    urls = [
        "file:///home/heath/Downloads/04784f4a830cc0df1f419a492cde9fc323f728db.tar.gz",
        "https://github.com/bazelbuild/remote-apis/archive/04784f4a830cc0df1f419a492cde9fc323f728db.tar.gz",
    ],
)
load("@bazel_remote_apis//:repository_rules.bzl", "switched_rules_by_language")
switched_rules_by_language(
    name = "bazel_remote_apis_imports",
    go = True,
)

http_archive(
    name = "com_github_grpc_grpc",
    sha256 = "c3e252efe5dc4ba30c0a2717b575d4fdd1fd9dbe62ec45b0920b4eb7fc28ef20",
    strip_prefix = "grpc-01f333a1c1a712eb347c3eb898d430693ca8e681",
    urls = ["https://github.com/grpc/grpc/archive/01f333a1c1a712eb347c3eb898d430693ca8e681.tar.gz"],
)

http_archive(
    name = "googleapis",
    build_file = "@bazel_remote_apis//:external/BUILD.googleapis",
    sha256 = "7b6ea252f0b8fb5cd722f45feb83e115b689909bbb6a393a873b6cbad4ceae1d",
    strip_prefix = "googleapis-143084a2624b6591ee1f9d23e7f5241856642f4d",
    urls = ["https://github.com/googleapis/googleapis/archive/143084a2624b6591ee1f9d23e7f5241856642f4d.zip"],
)
