dnf install -y rpm-build rpmdevtools
cd /mnt/nuv-$VER
rpmbuild --target=$TGT --buildroot=/mnt/nuv-$VER -bb /mnt/nuv.spec
