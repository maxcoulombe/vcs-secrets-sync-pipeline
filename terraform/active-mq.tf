resource "aws_security_group" "hack_week_active_mq" {
  name        = "hack_week_active_mq"

  ingress {
    from_port = 5671
    to_port   = 5671
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_mq_configuration" "hack_week_active_mq" {
  description    = "Example Configuration"
  name           = "hack_week"
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
  <plugins>
    <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
    <statisticsBrokerPlugin/>
    <timeStampingBrokerPlugin ttlCeiling="86400000" zeroExpirationOverride="86400000"/>
  </plugins>
</broker>
DATA
}

resource "aws_mq_broker" "hack_week_active_mq" {
  broker_name = "hack_week"

  configuration {
    id       = aws_mq_configuration.hack_week_active_mq.id
    revision = aws_mq_configuration.hack_week_active_mq.latest_revision
  }

  engine_type        = "ActiveMQ"
  engine_version     = "5.15.9"
  host_instance_type = "mq.t2.micro"
  security_groups    = [aws_security_group.hack_week_active_mq.id]

  user {
    username = "root"
    password = "Test4sk8board"
  }
}