#!/bin/bash

# -------------------------------

function create_new_db {
    Q00="CREATE DATABASE ${WP_DBNAME};"
    Q01="USE ${WP_DBNAME};"
    Q02="CREATE USER ${WP_DBUSER}@localhost IDENTIFIED BY '${WP_DBPASS}';"
    Q03="GRANT ALL PRIVILEGES ON ${WP_DBNAME}.* TO ${WP_DBUSER}@localhost;"
    Q04="FLUSH PRIVILEGES;"
    SQL0="${Q00}${Q01}${Q02}${Q03}${Q04}"
    mysql -v -u "root" -p${MYSQL_ROOT_PASSWD} -e"$SQL0"
}

function install_wp {
    wget http://wordpress.org/latest.tar.gz
    tar xzvf latest.tar.gz
    cp -rf wordpress/** ./
    rm -R wordpress
    cp wp-config-sample.php wp-config.php
    sed -i "s/database_name_here/${WP_DBNAME}/g" wp-config.php
    sed -i "s/username_here/${WP_DBUSER}/g" wp-config.php
    sed -i "s/password_here/${WP_DBPASS}/g" wp-config.php
    wget -O wp.keys https://api.wordpress.org/secret-key/1.1/salt/
    sed -i '/#@-/r wp.keys' wp-config.php
    sed -i "/#@+/,/#@-/d" wp-config.php
    mkdir wp-content/uploads
    find . -type d -exec chmod 755 {} \;
    find . -type f -exec chmod 644 {} \;
    chown -R :www-data wp-content/uploads
    chown -R $USER:www-data *
    chmod 640 wp-config.php
    rm -f latest.tar.gz
    rm -f wp.keys
    rm -f index.html
}

function generate_htaccess {
    touch .htaccess
    chown :www-data .htaccess
    chmod 644 .htaccess
    bash -c "cat > .htaccess" << _EOF_
# Block the include-only files.
<IfModule mod_rewrite.c>
RewriteEngine On
RewriteBase /
RewriteRule ^wp-admin/includes/ - [F,L]
RewriteRule !^wp-includes/ - [S=3]
RewriteRule ^wp-includes/[^/]+\.php$ - [F,L]
RewriteRule ^wp-includes/js/tinymce/langs/.+\.php - [F,L]
RewriteRule ^wp-includes/theme-compat/ - [F,L]
</IfModule>

# BEGIN WordPress
<IfModule mod_rewrite.c>
RewriteEngine On
RewriteBase /
RewriteRule ^index\.php$ - [L]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule . /index.php [L]
</IfModule>
# END WordPress

# Prevent viewing of .htaccess file
<Files .htaccess>
order allow,deny
deny from all
</Files>
# Prevent viewing of wp-config.php file
<files wp-config.php>
order allow,deny
deny from all
</files>
# Prevent directory listings
Options All -Indexes
_EOF_
}

function generate_robots {
    touch robots.txt
    bash -c "cat > robots.txt" << _EOF_
# Sitemap: absolute url
User-agent: *
Disallow: /cgi-bin/
Disallow: /wp-admin/
Disallow: /wp-includes/
Disallow: /wp-content/plugins/
Disallow: /wp-content/cache/
Disallow: /wp-content/themes/
Disallow: /trackback/
Disallow: /comments/
Disallow: */trackback/
Disallow: */comments/
Disallow: wp-login.php
Disallow: wp-signup.php
_EOF_
}

function download_plugins {
    cd wp-content/plugins/
    # W3 Total Cache
    plugin_url=$(curl -s https://wordpress.org/plugins/w3-total-cache/ | egrep -o "https://downloads.wordpress.org/plugin/[^']+")
    wget $plugin_url
    # Theme Test Drive
    plugin_url=$(curl -s https://wordpress.org/plugins/theme-test-drive/ | egrep -o "https://downloads.wordpress.org/plugin/[^']+")
    wget $plugin_url
    # Login LockDown
    plugin_url=$(curl -s https://wordpress.org/plugins/login-lockdown/ | egrep -o "https://downloads.wordpress.org/plugin/[^']+")
    wget $plugin_url
    # Easy Theme and Plugin Upgrades
    plugin_url=$(curl -s https://wordpress.org/plugins/easy-theme-and-plugin-upgrades/ | egrep -o "https://downloads.wordpress.org/plugin/[^']+")
    wget $plugin_url
    # Install unzip package
    apt-get install unzip
    # Unzip all zip files
    unzip \*.zip
    # Remove all zip files
    rm -f *.zip
    echo ""
    cd ../..
}

function post_process {
    curl -X POST -d "${2}" "${1}"
}

# -------------------------------

MYSQL_ROOT_PASSWD=123
WP_DBNAME=wp_db
WP_DBUSER=wp
WP_DBPASS=456
broadcast=http://10.140.0.2/ws-broadcast

# 1
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"apt-get update"}'

apt-get update

# 2
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"install apache php"}'

apt-get install -y apache2 php5 php5-mysql php5-curl php5-gd

systemctl restart apache2


export DEBIAN_FRONTEND="noninteractive"

debconf-set-selections <<< "mysql-server mysql-server/root_password password ${MYSQL_ROOT_PASSWD}"
debconf-set-selections <<< "mysql-server mysql-server/root_password_again password ${MYSQL_ROOT_PASSWD}"

# 3
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"install mysql"}'

apt-get install -y mysql-client mysql-server

# 4
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"setup html dir"}'

chown -R www-data:www-data /var/www/html/
cd /var/www/html/

# 5
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"create database"}'
create_new_db >> /tmp/debug

# 6
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"install WordPress"}'
install_wp >> /tmp/debug

# 7
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"generate .htaccess"}'
generate_htaccess >> /tmp/debug

generate_robots >> /tmp/debug

# 8
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"download WordPress plugins"}'
download_plugins >> /tmp/debug

# 9
post_process $broadcast '{"hostname":"${HOSTNAME}","active":"all done"}'


