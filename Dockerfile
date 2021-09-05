FROM scratch

WORKDIR /workspace
ADD paste paste
ADD index.html index.html

ENTRYPOINT ["./paste"]
