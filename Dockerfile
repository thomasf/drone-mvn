# Docker image for the Drone build runner
#

FROM java:8

RUN mkdir -p /opt \
      && cd /opt \
      && wget -q http://apache.mirrors.spacedump.net/maven/maven-3/3.3.3/binaries/apache-maven-3.3.3-bin.tar.gz \
      && tar -xf apache-maven-3.3.3-bin.tar.gz \
      && ln -s apache-maven-3.3.3 apache-maven

ENV PATH=/opt/apache-maven/bin:$PATH

# hacky way to get most maven dependencies into the cache
run echo '<settings></settings>' > /tmp/s \
    && mvn -q -s /tmp/s \
    deploy:deploy-file \
    -Durl=file:/tmp/t \
    -DrepositoryId=t \
    -Dfile=/bin/ls \
    -DartifactId=t \
    -Dversion=1.0 \
    -DgroupId=t \
    -Dpackaging=t \
    && rm -rf /tmp/t \
    && rm -rf /tmp/s \
    && rm -rf /root/.m2/repository/t

run echo '<settings></settings>' > /tmp/s \
    && mvn -q -s /tmp/s \
    gpg:sign-and-deploy-file \
    -Durl=file:/tmp/t \
    -DrepositoryId=t \
    -Dfile=/bin/ls \
    -DartifactId=t \
    -Dversion=1.0 \
    -DgroupId=t \
    -Dpackaging=t \
    || rm -rf /tmp/t \
    && rm -rf /tmp/s \
    && rm -rf /root/.m2/repository/t


ADD drone-mvn /bin/
ENTRYPOINT ["/bin/drone-mvn"]
