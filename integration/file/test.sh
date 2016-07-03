remco poll --onetime  file \
    --log-level=debug \
    --src=./integration/templates/basic.conf.tmpl \
    --dst=/tmp/remco-basic-test.conf \
    --filepath=./integration/file/config.yml \

cmp /tmp/remco-basic-test.conf ./integration/config/test.config || cat /tmp/remco-basic-test.conf