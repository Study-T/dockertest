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

        stage('运行测试') {
            agent {
                docker {
                    image 'golang:1.25-alpine'
                    args '--network host'
                }
            }
            steps {
                unstash 'source'
                sh '''
                    apk add --no-cache git gcc musl-dev
                    export GOWORK=$(pwd)/go.work
                    go work sync
                    cd app && go test ./... -v -count=1 && cd ..
                    cd domain/tracking && go test ./... -v -count=1 && cd ..
                    cd infrastructure && go test ./... -v -count=1 && cd ..
                    cd pkg && go test ./... -v -count=1
                '''
            }
        }

        stage('构建 Docker 镜像') {
            agent any
            steps {
                unstash 'source'
                sh 'docker build -t tracking-api:latest .'
            }
        }

        stage('部署到本地 Docker') {
            agent any
            steps {
                sh '''
                    docker compose down || true
                    docker compose up -d --build
                '''
            }
        }

        stage('运行数据库迁移') {
            agent any
            steps {
                sh '''
                    sleep 5
                    docker cp infrastructure/database/migration/001_create_raw_events.sql tracking-postgres:/tmp/migration.sql
                    docker exec tracking-postgres psql -U postgres -d your_database -f /tmp/migration.sql
                '''
            }
        }

        stage('健康检查') {
            agent any
            steps {
                sh '''
                    sleep 10
                    docker exec tracking-api wget -qO- http://localhost:8082/health || echo "健康检查失败，查看日志: docker logs tracking-api"
                    echo "==> 应用已成功部署，访问地址: http://localhost:8082"
                '''
            }
        }
    }

    post {
        success {
            echo '✅ 流水线执行成功！Go 应用已部署到本地 Docker'
        }
        failure {
            echo '❌ 流水线执行失败，请检查控制台输出'
        }
    }
}
