#!/bin/bash

DEST=$1

VERSION=$(cat VERSION)
PKGVERSION="${VERSION:-latest}"
PACKAGE_ARCHITECTURE="${ARCHITECTURE:-amd64}"
PACKAGE_URL="http://www.tutum.co/"
PACKAGE_MAINTAINER="support@tutum.co"
PACKAGE_DESCRIPTION="Agent to manage Docker hosts through Tutum"
PACKAGE_LICENSE="Proprietary"

bundle_redhat() {
  DIR=$DEST/staging

  # Include our init scripts
  mkdir -p $DIR/etc/init
  mkdir -p $DIR/etc/init.d
  cp contrib/upstart/tutum-agent.conf $DIR/etc/init/
  cp -v contrib/init.d/tutum-agent $DIR/etc/init.d/

  # Copy the binary
  # This will fail if the binary bundle hasn't been built
  mkdir -p $DIR/usr/bin
  cp -p /build/bin/linux/$PACKAGE_ARCHITECTURE/tutum-agent-$PKGVERSION $DIR/usr/bin/tutum-agent

  cat > $DEST/postinst <<'EOF'
#!/bin/sh
set -e

DOCKER_UPSTART_CONF="/etc/init/docker.conf"
if [ -f "${DOCKER_UPSTART_CONF}" ]; then
  echo "Removing conflicting docker upstart configuration file at ${DOCKER_UPSTART_CONF}..."
  rm -f ${DOCKER_UPSTART_CONF}
fi

if [ -n "$2" ]; then
  _dh_action=restart
else
  _dh_action=start
fi
service tutum-agent $_dh_action 2>/dev/null || true

#DEBHELPER#
EOF

  cat > $DEST/prerm <<'EOF'
#!/bin/sh
set -e

case "$1" in
  remove)
    service tutum-agent stop 2>/dev/null || true
  ;;
esac

#DEBHELPER#
EOF

  cat > $DEST/postrm <<'EOF'
#!/bin/sh
set -e

case "$1" in
  remove)
    rm -fr /usr/bin/docker /usr/lib/tutum
  ;;
esac

#DEBHELPER#
EOF

  chmod +x $DEST/postinst $DEST/prerm $DEST/postrm

  (
    # switch directories so we create *.deb in the right folder
    cd $DEST

    apt-get install -y rpm
    # create tutum-agent-$PKGVERSION package
    fpm -s dir -C $DIR \
      --name tutum-agent --version $PKGVERSION \
      --after-install $DEST/postinst \
      --before-remove $DEST/prerm \
      --after-remove $DEST/postrm \
      --architecture "$PACKAGE_ARCHITECTURE" \
      --prefix / \
      --description "$PACKAGE_DESCRIPTION" \
      --maintainer "$PACKAGE_MAINTAINER" \
<<<<<<< HEAD
      -t rpm . \
      --conflicts docker \
      --conflicts docker.io \
      --conflicts lxc-docker \
      --depends cgroup-lite \
      --depends aufs-tools \
      --depends iptables \
      --depends "libapparmor1 >= 2.6~devel" \
      --depends "libc6 >= 2.4" \
      --depends "libdevmapper1.02.1 >= 2:1.02.63" \
      --depends "libsqlite3-0 >= 3.5.9" \
      --depends perl \
      --depends gnupg \
      --depends "sysv-rc >= 2.88dsf-24" \
      --depends xz-utils \
      --config-files "etc/init/tutum-agent.conf" \
      --config-files "etc/init.d/tutum-agent" \
      --conflicts docker \
      --conflicts docker.io \
      --conflicts lxc-docker \
      --depends iptables \
      --depends "device-mapper-libs" \
      --depends perl \
      --depends gnupg \
      --depends xz \
      --provides tutum-agent \
      --replaces tutum-agent \
      --url "$PACKAGE_URL" \
      --license "$PACKAGE_LICENSE" \
      -t rpm .
  )

  rm $DEST/postinst $DEST/prerm $DEST/postrm
  rm -r $DIR
}

bundle_redhat
