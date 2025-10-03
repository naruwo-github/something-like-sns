#  アプリケーション全体が動作するための隔離されたネットワーク空間（VPC）と、その通信ルールを定義します。
#
#   * aws_vpc (1個): クラウド内のプライベートなネットワーク全体です。
#   * aws_subnet (計4個): VPCをさらに細かく分割したネットワーク。インターネットに接続できるパブリックサブネット2つと、直接接続できないプライベートサブネット2つを作成しています。
#   * aws_internet_gateway (1個): VPCをインターネットに接続するための出入り口です。
#   * aws_eip (1個): この後作成するNATゲートウェイに固定のグローバルIPアドレスを割り当てるためのものです。
#   * aws_nat_gateway (1個): プライベートサブネット内のリソースが、外のインターネットにアクセス（例: OSのアップデートなど）するために使う出口です。
#   * aws_route_table (計2個): 通信の経路を決めるルートテーブル。パブリック用とプライベート用の2つです。
#   * aws_route_table_association (計4個): 上記のルートテーブルを各サブネット（計4つ）に関連付けます。
#
#  【このファイルの合計: 14リソース】

# AZ情報を動的に取得するためのlocal変数を定義 (最初の2つを利用)
locals {
  az_names = slice(data.aws_availability_zones.available.names, 0, 2)
}

# 利用可能なアベイラビリティゾーンの情報を取得
data "aws_availability_zones" "available" {
  state = "available"
}

# メインのVPCを作成
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# パブリックサブネットを2つ作成
resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone       = local.az_names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.project_name}-public-subnet-${count.index + 1}"
  }
}

# プライベートサブネットを2つ作成
resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 2)
  availability_zone = local.az_names[count.index]

  tags = {
    Name = "${var.project_name}-private-subnet-${count.index + 1}"
  }
}

# インターネットゲートウェイを作成
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# NATゲートウェイ用のElastic IPを作成
resource "aws_eip" "nat" {
  domain = "vpc"
}

# NATゲートウェイを作成 (プライベートサブネットからのアウトバウンド通信用)
resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[0].id

  tags = {
    Name = "${var.project_name}-nat-gw"
  }

  depends_on = [aws_internet_gateway.main]
}

# パブリックルートテーブルを作成 (インターネットへのルート)
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-public-rt"
  }
}

# パブリックサブネットをルートテーブルに関連付け
resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# プライベートルートテーブルを作成 (NATゲートウェイへのルート)
resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-private-rt"
  }
}

# プライベートサブネットをルートテーブルに関連付け
resource "aws_route_table_association" "private" {
  count          = 2
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
