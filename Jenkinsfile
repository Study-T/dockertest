pipeline {
    agent any

    stages {
        stage('检出代码') {
            steps {
                echo '==> 开始拉取代码'
                checkout scm
            }
        }

        stage('测试') {
            steps {
                echo '==> 运行测试'
                sh 'echo "Hello from Jenkins CI!"'
            }
        }

        stage('构建') {
            steps {
                echo '==> 构建项目'
                sh 'echo "Build success!"'
            }
        }
    }

    post {
        success {
            echo '✅ 流水线执行成功'
        }
        failure {
            echo '❌ 流水线执行失败'
        }
    }
}
