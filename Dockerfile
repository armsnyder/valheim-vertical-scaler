# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details

FROM gcr.io/distroless/static:nonroot AS amd64
WORKDIR /
COPY bin/manager_linux_amd64 /manager
USER 65532:65532

ENTRYPOINT ["/manager"]


FROM gcr.io/distroless/static:nonroot AS arm64
WORKDIR /
COPY bin/manager_linux_arm64 /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
