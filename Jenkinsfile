pipeline {
    agent any
    stages {
        stage('backup slug') {
            when {
			    not {
                    branch 'master'
                    }
		    }
            steps {
				checkout scm
				script {
                    sh "rm -rf slug.tgz"
                    sh "aws s3 cp s3://agora-app-builder-backend-go-builds/slug.tgz ."
                    sh "mv slug.tgz previous_slug.tgz"
                    sh "aws s3 cp previous_slug.tgz s3://agora-app-builder-backend-go-builds/previous_slug.tgz"
                    sh "rm -rf slug.tgz"
                    sh "rm -rf previous_slug.tgz"
                    }
                }
            }

        stage('build slug') {
            when {
                not {
                    branch 'master'
                    }
            }
            steps {
				checkout scm
				script {
                    sh "pwd"
					sh " chmod 777 scripts/create_slugs.sh"
					sh " ls -la"
					sh "scripts/create_slugs.sh"
                    sh " ls -la"
                    }
                }
            }


		stage('push to s3') {
            when {
                not {
                    branch 'master'
                    }
            }
            steps {
                script {
                // sh "aws s3api put-object --bucket agora-app-builder-backend-go-builds --key ${BUILD_NUMBER}"
                sh "aws s3 cp slug.tgz s3://agora-app-builder-backend-go-builds/slug.tgz"
                }
            deleteDir()
            }
		}
    }
}
