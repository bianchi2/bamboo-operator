package deploy

import (
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
)

func GetInstallBambooConfigMap(bamboo *installv1alpha1.Bamboo) *apiv1.ConfigMap {
	installBamboo := `
import requests
import time
import os
import re
import urllib.request

# grab Bamboo environment variables
protocol = os.getenv('PROTOCOL')
url = os.getenv('BAMBOO_ENDPOINT')
key = os.getenv('BAMBOO_LICENSE')
adminUser = os.getenv('ADMIN_USER')
adminPassword = os.getenv('ADMIN_PASSWORD')
email = os.getenv('ADMIN_EMAIL')
fullName = os.getenv('FULL_NAME')
bambooDbhost = os.getenv('BAMBOO_DATABASE_HOST')
bambooDbPort = os.getenv('BAMBOO_DATABASE_PORT')
bambooDbName = os.getenv('BAMBOO_DATABASE_NAME')
bambooDbUser = os.getenv('BAMBOO_DATABASE_USER')
bambooDbPassword = os.getenv('BAMBOO_DATABASE_PASSWORD')

print('Bamboo installer welcomes you')
connection = None
while connection == None:
    try:
        connection = requests.get(protocol + '://' + url)
    except:
        print("Bamboo hostname cannot be yet resolved")
        time.sleep(5)
response = requests.get(protocol + '://' + url)
if response.status_code == 200:
    print('Bamboo is ready, response code:', response.status_code)
# wait for Bamboo to start responding
while response.status_code != 200:
    response = requests.get(protocol + '://' + url)
    print('Bamboo is not ready, response code:', response.status_code)
    time.sleep(5)

selectSetupStep = urllib.request.urlopen(protocol + '://' + url + '/bootstrap/selectSetupStep.action')
cookies = selectSetupStep.headers

m = re.search('atl.xsrf.token=(.+?);', str(cookies))
atl_token = m.group(1)
print('Using auth token: ' + atl_token)
licenseData = {'sid': 'B4JP-MO66-CRS9-E2MH',
               'licenseString': key,
               'customInstall': 'Custom+installation',
               'atl_token': atl_token
               }
headers = {'Cookie': 'atl.xsrf.token=' + atl_token + ';'}
print('Validating license')
validateLicense = requests.post(protocol + '://' + url + '/setup/validateLicense.action', data=licenseData, headers=headers)
time.sleep(5)

dbData = {
            'dbChoice': 'standardDb',
            'selectedDatabase': 'postgresql',
            'selectFields': 'selectedDatabase',
            'save': 'Continue',
            'atl_token': atl_token
}
chooseDatabase = requests.post(protocol + '://' + url + '/setup/chooseDatabaseType.action', data=dbData, headers=headers)

time.sleep(10)


pgData = {
            'selectedDatabase': 'postgresql',
            'connectionChoice': 'jdbcConnection',
            'dbConfigInfo.driverClassName': 'org.postgresql.Driver',
            'dbConfigInfo.databaseUrl': 'jdbc:postgresql://' + bambooDbhost + ':' + bambooDbPort + '/' + bambooDbName,
            'dbConfigInfo.userName': bambooDbUser,
            'dbConfigInfo.password': bambooDbPassword,
            'checkBoxFields=data': 'dataOverwrite',
            'dataOverwrite': 'true',
            'atl_token': atl_token
}
print('Setting up database')
setupDatabase = requests.post(protocol + '://' + url + '/setup/performSetupDatabaseConnection.action', data=pgData, headers=headers)

time.sleep(10)
# wait for Bamboo to be ready to create admin user. timeout == 5 mins

nextStep = requests.get(protocol + '://' + url + '/bootstrap/selectSetupStep.action')

timeout = time.time() + 60*5
while nextStep.history[0].headers['Location'] != '/setup/setupAdminUser.action':
    if time.time() > timeout:
        print('Timeout reached. Bamboo is not ready to create a user. Check logs')
        sys.exit('Exiting')
    nextStep = requests.get(protocol + '://' + url + '/bootstrap/selectSetupStep.action')
    print('Suggested redirect is ' + nextStep.history[0].headers['Location'] + ' Waiting for /setup/setupAdminUser.action')
    time.sleep(15)

print('Creating admin user at ' + nextStep.history[0].headers['Location'])


adminUserData = {
            'username': adminUser,
            'password': adminPassword,
            'confirmPassword': adminPassword,
            'fullName': fullName,
            'email': email,
            'save': 'Finish',
            'atl_token': atl_token
}
print('Creating admin user')
createAdmUser = requests.post(protocol + '://' + url + '/setup/performSetupAdminUser.action', data=adminUserData, headers=headers)
print(createAdmUser)
print('Bamboo is installed')
`
	re := regexp.MustCompile(`\r?\n\\n`)
	installBamboo = re.ReplaceAllString(installBamboo, " ")

	configMapData := make(map[string]string, 0)
	configMapData["install-bamboo.py"] = installBamboo
	installBambooConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "install-bamboo-py",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	return installBambooConfigMap
}
