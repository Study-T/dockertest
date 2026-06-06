pipeline {
    agent none

    stages {
        stage('拉取代码') {
            agent any
            steps {
                echo '==> 拉取 GitHub 代码'
                checkout scm
                stash name: 'source', includes: '**/*', useDefaultExcludes: false
            }
        }

        stage('安装依赖 + 运行测试') {
            agent {
                docker {
                    image 'python:3.12-slim'
                    args '--network host'
                }
            }
            steps {
                unstash 'source'
                sh '''
                    pip install --no-cache-dir -r requirements.txt
                    pytest test_app.py -v
                '''
            }
        }

        stage('构建 Docker 镜像') {
            agent any
            steps {
                unstash 'source'
                sh 'docker build -t dockertest:latest .'
            }
        }

        stage('部署到本地 Docker') {
            agent any
            steps {
                sh '''
                    docker stop dockertest || true
                    docker rm dockertest || true
                    docker run -d \
                        --name dockertest \
                        -p 9090:8000 \
                        --restart unless-stopped \
                        dockertest:latest
                '''
            }
        }

        stage('健康检查') {
            agent any
            steps {
                sh '''
                    sleep 5
                    docker exec dockertest curl -sf http://localhost:8000/health || exit 1
                    echo ""
                    echo "==> 应用已成功部署，访问地址: http://localhost:9090"
                '''
            }
        }
    }

    post {
        success {
            echo '✅ 流水线执行成功！应用已部署到本地 Docker'
        }
        failure {
            echo '❌ 流水线执行失败，请检查控制台输出'
        }
    }
}