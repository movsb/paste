FROM scratch

WORKDIR /workspace
ADD paste paste
ENTRYPOINT ["./paste"]
