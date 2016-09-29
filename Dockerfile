FROM scratch
COPY ./lapinou /lapinou
ENTRYPOINT ["/lapinou"]
