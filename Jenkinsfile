pipeline {
    agent any
    stages {
        stage('backup slug-dev') {
            when {
                not{
                    branch 'master'
                }
		    }
            steps {
				checkout scm
				script {
                    sh "rm -rf slug.tgz || echo 'slug.tgz not present' "
                    sh "aws s3 cp s3://agora-app-builder-backend-go-builds/dev/slug.tgz . || echo 'slug.tgz not present' "
                    sh "mv slug.tgz slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    sh "aws s3 cp slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz s3://agora-app-builder-backend-go-builds/dev/slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    sh "rm -rf slug.tgz || echo 'slug.tgz not present' "
                    sh "rm -rf slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    }
                }
            }

        stage('backup slug-staging') {
            when {
                branch 'master'
		    }
            steps {
				checkout scm
				script {
                    sh "rm -rf slug.tgz || echo 'slug.tgz not present' "
                    sh "aws s3 cp s3://agora-app-builder-backend-go-builds/staging/slug.tgz . || echo 'slug.tgz not present' "
                    sh "mv slug.tgz slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    sh "aws s3 cp slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz s3://agora-app-builder-backend-go-builds/staging/slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    sh "rm -rf slug.tgz || echo 'slug.tgz not present' "
                    sh "rm -rf slug_${BUILD_NUMBER}_${BUILD_TIMESTAMP}.tgz || echo 'slug.tgz not present' "
                    }
                }
            }

        stage('build slug') {
            steps {
				checkout scm
				script {
                    sh "pwd"
					sh "chmod 777 scripts/create_slugs.sh"
					sh "ls -la"
					sh "scripts/create_slugs.sh"
                    sh "ls -la"
                    sh "ls -la ./app/"
                    }
                }
            }


		stage('push to s3-dev') {
            when {
                not{
                    branch 'master'
                }
		    }
            steps {
                script {
                    sh "aws s3 cp slug.tgz s3://agora-app-builder-backend-go-builds/dev/slug.tgz"
                }
            deleteDir()
            }
		}

		stage('push to s3-staging') {
            when {
                branch 'master'
		    }
            steps {
                script {
                    sh "aws s3 cp slug.tgz s3://agora-app-builder-backend-go-builds/staging/slug.tgz"
                }
            deleteDir()
            }
		}
    }
}
