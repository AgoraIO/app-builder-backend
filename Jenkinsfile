pipeline {
    agent any
    stages {
        stage('build slug') {
		when {
			branch "feature/*"
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
			branch "feature/*"
		}
		steps {
			script {
                sh "pwd"
                sh "aws s3api put-object --bucket agora-app-builder-backend-go-builds --key ${BUILD_NUMBER}/"
                sh "aws s3 cp slug.tgz s3://agora-app-builder-backend-go-builds/${BUILD_NUMBER}"
			}
            deleteDir()
		}
		}

    }
}
