package deploy

import (
	"fmt"
	installv1alpha1 "github.com/bianchi2/bamboo-operator/api/v1alpha1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func GetLoggingPropertiesConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	logginProperties := `
#
# Change the following line to configure the bamboo logging levels (one of INFO, DEBUG, ERROR, FATAL)
#
log4j.rootLogger=INFO, console
log4j.appender.console=org.apache.log4j.ConsoleAppender
log4j.appender.console.layout=org.apache.log4j.PatternLayout
log4j.appender.console.layout.ConversionPattern=%d %p [%t] [%c{1}] %m%n
# log4j.appender.console.threshold = OFF

#using 'bamboo home aware' appender. If the File is relative a relative Path the file goes into {bamboo.home}/logs
# log4j.appender.filelog=com.atlassian.bamboo.log.BambooRollingFileAppender
# log4j.appender.filelog.File=atlassian-bamboo.log
# log4j.appender.filelog.MaxFileSize=100MB
# log4j.appender.filelog.MaxBackupIndex=5
# log4j.appender.filelog.layout=org.apache.log4j.PatternLayout
# log4j.appender.filelog.layout.ConversionPattern=%d %p [%t] [%c{1}] %m%n

#using 'bamboo home aware' appender for access logs
log4j.appender.accesslog=com.atlassian.bamboo.log.BambooRollingFileAppender
log4j.appender.accesslog.File=atlassian-bamboo-access.log
log4j.appender.accesslog.MaxFileSize=20480KB
log4j.appender.accesslog.MaxBackupIndex=5
log4j.appender.accesslog.layout=org.apache.log4j.PatternLayout
log4j.appender.accesslog.layout.ConversionPattern=%d %p [%t] [%c{1}] %m%n

# This log below gives more correct line can class details
#log4j.appender.console.layout.ConversionPattern=%d %p [%t] [%C{1}:%L] %m%n
#log4j.appender.filelog.layout.ConversionPattern=%d %p [%t] [%C{1}:%L] %m%n


########################################################################################################################
# Access Log
########################################################################################################################

log4j.category.com.atlassian.bamboo.filter.AccessLogFilter=INFO, accesslog
log4j.additivity.com.atlassian.bamboo.filter.AccessLogFilter=false


########################################################################################################################
# Database Logging
########################################################################################################################

log4j.category.org.hibernate.impl.SessionImpl=ERROR
## log hibernate prepared statements/SQL queries (equivalent to setting 'hibernate.show_sql' to 'true')
#log4j.category.org.hibernate.SQL=DEBUG
## log hibernate prepared statement parameter values
#log4j.category.org.hibernate.type=TRACE
#log4j.category.org.hibernate.impl.BatcherImpl=DEBUG
## log active objects sql statements
#log4j.category.net.java.ao.sql=DEBUG
## Bamboo plan cache detailed logging
# log4j.category.com.atlassian.bamboo.plan.cache = DEBUG
#Remove when criteria queries are mostly gone
log4j.logger.org.hibernate.orm.deprecation=ERROR
#Remove when most entities are on annotations instead of hbms
log4j.logger.org.hibernate.metamodel.internal.MetadataContext=ERROR

########################################################################################################################
# OSGi Related logs
########################################################################################################################

log4j.category.com.atlassian.plugin.osgi.container.felix.FelixOsgiContainerManager=WARN
log4j.category.org.twdata.pkgscanner.ExportPackageListBuilder=ERROR
log4j.category.com.atlassian.plugin=WARN
log4j.category.com.atlassian.plugin.osgi=WARN
log4j.category.com.atlassian.streams=WARN

#log4j.logger.com.atlassian.bamboo.plugins.atlassian-bamboo-plugin-sal.spring  = DEBUG
#log4j.logger.com.atlassian.bamboo.plugins.bamboo-activeobjects-spi.spring  = DEBUG
#log4j.logger.com.atlassian.activeobjects.activeobjects-plugin.spring  = DEBUG
#log4j.logger.com.atlassian.plugins.rest.atlassian-rest-module.spring  = DEBUG


########################################################################################################################
# SSH proxy related settings
########################################################################################################################

log4j.category.org.apache.sshd=WARN
log4j.category.com.atlassian.bamboo.plugins.hg.sshproxy=WARN
log4j.category.com.atlassian.bamboo.plugins.ssh=WARN
log4j.category.com.atlassian.bamboo.ssh=WARN



########################################################################################################################
# Queue and agent related
########################################################################################################################

#log4j.category.com.atlassian.bamboo.agent=DEBUG
#log4j.category.com.atlassian.bamboo.buildqueue=DEBUG
#log4j.category.com.atlassian.bamboo.v2.build.agent.remote.heartbeat.AgentHeartBeatJob=DEBUG
#log4j.category.com.atlassian.bamboo.buildqueue.manager.RemoteAgentManagerImpl=DEBUG



########################################################################################################################
# For task process env + details
########################################################################################################################

#log4j.category.com.atlassian.utils.process=DEBUG

# task execution exception details logging
#log4j.logger.com.atlassian.bamboo.executor=DEBUG


########################################################################################################################
# Repository Debugging
########################################################################################################################

#log4j.logger.com.atlassian.stash.rest.client.core.StashClientImpl=TRACE
#log4j.logger.com.atlassian.bamboo.plugins.hg.HgRepository=DEBUG
#log4j.logger.com.atlassian.bamboo.plugins.git.GitRepository=DEBUG
#log4j.logger.com.perforce=DEBUG
#log4j.logger.com.tek42=DEBUG

########################################################################################################################
# Emergency, high volume logging
########################################################################################################################

log4j.logger.Emergency=INFO, emergency
log4j.additivity.Emergency=false
log4j.appender.emergency=com.atlassian.bamboo.log.BambooRollingFileAppender
log4j.appender.emergency.File=emergency-atlassian-bamboo.log
log4j.appender.emergency.MaxFileSize=20480KB
log4j.appender.emergency.MaxBackupIndex=5
log4j.appender.emergency.layout=org.apache.log4j.PatternLayout
log4j.appender.emergency.layout.ConversionPattern=%d %p [%t] [%c{1}] %m%n

########################################################################################################################
# Javascript Debugging
########################################################################################################################

log4j.logger.JavaScript=DEBUG, javascript
log4j.additivity.JavaScript=false
log4j.appender.javascript=com.atlassian.bamboo.log.BambooRollingFileAppender
log4j.appender.javascript.File=js-atlassian-bamboo.log
log4j.appender.javascript.MaxFileSize=20480KB
log4j.appender.javascript.MaxBackupIndex=5
log4j.appender.javascript.layout=org.apache.log4j.PatternLayout
log4j.appender.javascript.layout.ConversionPattern=%d %p [%t] [%c{1}] %m%n

########################################################################################################################
# Other random debugging
########################################################################################################################

#log4j.logger.com.atlassian.bamboo.jira.issuelink=DEBUG
#log4j.logger.com.atlassian.bamboo.v2.build.agent.messages=DEBUG


########################################################################################################################
# General ignore logging for various packages
########################################################################################################################

log4j.category.com.atlassian.marketplace.client.MarketplaceClient=WARN
log4j.category.webwork=WARN
log4j.category.org.apache.velocity=WARN
log4j.category.com.opensymphony.xwork2.util.LocalizedTextUtil=ERROR
log4j.category.com.opensymphony.xwork2.util.OgnlValueStack=ERROR
log4j.category.org.springframework.beans.factory.support.DependencyInjectionAspectSupport=WARN
log4j.category.org.springframework=WARN
log4j.category.org.hibernate=WARN
log4j.category.org.acegisecurity=WARN
log4j.category.org.apache.activemq.transport.failover.FailoverTransport=WARN
log4j.category.com.atlassian.plugin.servlet.filter.ServletFilterModuleContainerFilter=WARN
log4j.category.com.atlassian.plugin.servlet.PluginResourceDownload=WARN
log4j.category.com.atlassian.bamboo.persister.xstream.IgnoreMissingFieldXStream=INFO
log4j.category.com.atlassian.crowd=INFO
log4j.category.com.atlassian.plugin.webresource.DefaultResourceDependencyResolver=ERROR
log4j.logger.org.apache=INFO
log4j.logger.org.eclipse.jetty=INFO
log4j.logger.cz.vutbr.web=WARN
log4j.logger.com.atlassian.botocss=WARN
log4j.logger.com.amazonaws=WARN
# log S3 retries
log4j.logger.com.amazonaws.services.s3=INFO
log4j.logger.org.apache.struts2.config.BeanSelectionProvide=WARN
log4j.logger.com.atlassian.tunnel=WARN
# Docker client
log4j.logger.com.spotify.docker.client = WARN

# user management
log4j.category.com.atlassian.bamboo.user.BambooUserManagerImpl=WARN
# Embedded Crowd
log4j.logger.com.atlassian.crowd=WARN

# {{ if .Values.tls.use }}https{{ else }}http{{ end }}://jira.atlassian.com/browse/ROTP-1546 bumping logging level for userprovisioning plugin to see information about group migration in Cloud when upgrading to unified user management
# this can be removed once we have switched all instances over to unified user management
log4j.logger.com.atlassian.usermanagement.userprovisioning=INFO

# debugging immutable plan cache
#log4j.category.com.atlassian.bamboo.plan.cache.ImmutablePlanCacheServiceImpl=DEBUG
#log4j.category.com.atlassian.bamboo.plan.cache.CacheLoadContextSupport=DEBUG

# debug webhooks
#log4j.category.com.atlassian.bamboo.plugins.jira.event=TRACE

# profiling (needs system property -Datlassian.profile.activate=true to be activated)
log4j.category.com.atlassian.util.profiling=DEBUG

# query execution times
#log4j.logger.com.atlassian.bamboo.utils.db.JdbcUtils = ALL

#upm cache flushes
log4j.category.com.atlassian.cache.stacktrace=WARN
`
	re := regexp.MustCompile(`\r?\n\\n`)
	logginProperties = re.ReplaceAllString(logginProperties, " ")

	configMapData := make(map[string]string, 0)
	configMapData["logging.properties"] = logginProperties
	loggingPropertiesConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bamboo-logging-properties",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, loggingPropertiesConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return loggingPropertiesConfigMap
}

func GetAdministrationXMLConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	protocol := "http"
	if bamboo.Spec.Ingress.Tls {
		protocol = "https"
	}
	adminXml := `<AdministrationConfiguration>
<myBaseUrl>` + protocol + `://` + bamboo.Spec.Ingress.Host + `</myBaseUrl>
<remoteAgentFunctionEnabled>true</remoteAgentFunctionEnabled>
</AdministrationConfiguration>
`
	re := regexp.MustCompile(`\r?\n\\n`)
	adminXml = re.ReplaceAllString(adminXml, " ")

	configMapData := make(map[string]string, 0)
	configMapData["administration.xml"] = adminXml
	administrationXMLConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bamboo-administration-xml",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, administrationXMLConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return administrationXMLConfigMap
}

func GetBambooAgentCfgConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	cfgXml := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<configuration>
<buildWorkingDirectory>/var/atlassian/application-data/bamboo-agent/xml-data/build-dir</buildWorkingDirectory>
<agentUuid>UID</agentUuid>
<agentDefinition>
<id>ID</id>
<name>remote-agent-NAME</name>
<description>Remote agent on host NAME</description>
</agentDefinition>
</configuration>
`
	re := regexp.MustCompile(`\r?\n\\n`)
	cfgXml = re.ReplaceAllString(cfgXml, " ")

	configMapData := make(map[string]string, 0)
	configMapData["bamboo-agent.cfg.xml"] = cfgXml
	cfgXMLConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bamboo-agent-cfg-xml",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, cfgXMLConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return cfgXMLConfigMap
}

func GetBambooCfgConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	cfgXml := `<?xml version="1.0" encoding="UTF-8"?>
<application-configuration>
  <setupStep>setupDatabase</setupStep>
  <setupType>install</setupType>
  <buildNumber>BUILD_NUMBER</buildNumber>
  <properties>
  <property name="bamboo.artifacts.directory">${bambooHome}/artifacts</property>
  <property name="bamboo.config.directory">${bambooHome}/xml-data/configuration</property>
  <property name="bamboo.jms.broker.client.uri">failover:(tcp://` + bamboo.Name + `:54663?wireFormat.maxInactivityDuration=300000)?maxReconnectAttempts=10&amp;initialReconnectDelay=15000</property>
  <property name="bamboo.jms.broker.uri">nio://0.0.0.0:54663</property>
  <property name="bamboo.project.directory">${bambooHome}/xml-data/builds</property>
  <property name="bamboo.repository.logs.directory">${bambooHome}/xml-data/repository-specs</property>
  <property name="buildWorkingDir">${bambooHome}/xml-data/build-dir</property>
  <property name="hibernate.c3p0.acquire_increment">3</property>
  <property name="hibernate.c3p0.idle_test_period">31</property>
  <property name="hibernate.c3p0.max_size">100</property>
  <property name="hibernate.c3p0.max_statements">0</property>
  <property name="hibernate.c3p0.min_size">3</property>
  <property name="hibernate.c3p0.timeout">130</property>
  <property name="lucene.index.dir">${bambooHome}/index</property>
  <property name="serverId">BRSG-A9XP-FCK3-NIML</property>
  <property name="serverKey">287</property>
  <property name="webwork.multipart.saveDir">${bambooHome}/temp</property>
  </properties>
</application-configuration>
`
	re := regexp.MustCompile(`\r?\n\\n`)
	cfgXml = re.ReplaceAllString(cfgXml, " ")

	configMapData := make(map[string]string, 0)
	configMapData["bamboo.cfg.xml"] = cfgXml
	cfgXMLConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bamboo-cfg-xml",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, cfgXMLConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return cfgXMLConfigMap
}

func GetBambooCreateConfigConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	bash := `
#!/bin/bash
FILE="/var/atlassian/application-data/bamboo/bamboo.cfg.xml"
if [ -f "$FILE" ]; then
	echo "$FILE exist and will not be overridden"
else
	mkdir -p /var/atlassian/application-data/bamboo/index/results /var/atlassian/application-data/bamboo/xml-data/configuration
	# copy xmls
	cp /tmp/bamboo.cfg.xml /var/atlassian/application-data/bamboo/bamboo.cfg.xml
	cp /tmp/administration.xml /var/atlassian/application-data/bamboo/xml-data/configuration/administration.xml
fi
export BUILD_NUMBER=$(curl -L --silent https://packages.atlassian.com/maven-external/com/atlassian/bamboo/atlassian-bamboo/` + bamboo.Spec.ImageTag + `/atlassian-bamboo-` + bamboo.Spec.ImageTag + `.pom | grep buildNumber | cut -d'>' -f 2| cut -d'<' -f 1)
sed -i 's/BUILD_NUMBER/'"$BUILD_NUMBER"'/' /var/atlassian/application-data/bamboo/bamboo.cfg.xml
`
	re := regexp.MustCompile(`\r?\t`)
	bash = re.ReplaceAllString(bash, " ")

	configMapData := make(map[string]string, 0)
	configMapData["create-config.sh"] = bash
	createConfigConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-config-sh",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, createConfigConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return createConfigConfigMap
}

func GetBambooCreateAgentConfigConfigMap(bamboo *installv1alpha1.Bamboo, bambooAPI BambooAPI) *apiv1.ConfigMap {
	bash := `
#!/bin/bash
FILE="/var/atlassian/application-data/bamboo-agent/bamboo-agent.cfg.xml"
if [ -f "$FILE" ]; then
	echo "$FILE exist and will not be overridden"
else
	# copy xml
	cp /tmp/bamboo-agent.cfg.xml /var/atlassian/application-data/bamboo-agent/bamboo-agent.cfg.xml
fi
export BUILD_NUMBER=$(curl -L --silent https://packages.atlassian.com/maven-external/com/atlassian/bamboo/atlassian-bamboo/` + bamboo.Spec.ImageTag + `/atlassian-bamboo-` + bamboo.Spec.ImageTag + `.pom | grep buildNumber | cut -d'>' -f 2| cut -d'<' -f 1)
sed -i "s/UID/${UID}/g" /var/atlassian/application-data/bamboo-agent/bamboo-agent.cfg.xml
sed -i "s/ID/${ID}/g" /var/atlassian/application-data/bamboo-agent/bamboo-agent.cfg.xml
sed -i "s/NAME/${ID}/g" /var/atlassian/application-data/bamboo-agent/bamboo-agent.cfg.xml
`
	re := regexp.MustCompile(`\r?\t`)
	bash = re.ReplaceAllString(bash, " ")

	configMapData := make(map[string]string, 0)
	configMapData["create-agent-config.sh"] = bash
	createConfigConfigMap := &apiv1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-agent-config-sh",
			Namespace: bamboo.Namespace,
		},
		Data: configMapData,
	}
	err := controllerutil.SetControllerReference(bamboo, createConfigConfigMap, bambooAPI.Scheme)
	if err != nil {
		fmt.Printf("An error occurred when setting controller reference: %s", err)
	}
	return createConfigConfigMap
}
