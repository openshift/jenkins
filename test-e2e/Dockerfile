FROM openshift/origin
RUN mkdir -p /usr/libexec/origin
COPY tag-in-image.sh /usr/libexec/origin
RUN echo "Copied tag-in-image.sh to /usr/libexec/origin"
RUN echo "looking into /usr/libexec/origin"
RUN ls -la /usr/libexec/origin