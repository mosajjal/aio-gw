# build statically to out/

DOCKER_BUILDKIT=1 docker build --file Dockerfile --output out .