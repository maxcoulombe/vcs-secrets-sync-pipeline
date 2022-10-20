#resource "aws_api_gateway_rest_api" "hack_week_gateway" {
#  name        = "hack-week-gateway"
#}
#
#resource "aws_api_gateway_resource" "hack_week_resource" {
#  rest_api_id = aws_api_gateway_rest_api.hack_week_gateway.id
#  parent_id   = aws_api_gateway_rest_api.hack_week_gateway.root_resource_id
#  path_part   = "{proxy+}"
#}
#
#resource "aws_api_gateway_method" "hack_week_proxy" {
#  rest_api_id   = aws_api_gateway_rest_api.hack_week_gateway.id
#  resource_id   = aws_api_gateway_resource.hack_week_resource.id
#  http_method   = "ANY"
#  authorization = "NONE"
#}
#
#resource "aws_api_gateway_integration" "hack_week_integration" {
#  rest_api_id = aws_api_gateway_rest_api.hack_week_gateway.id
#  resource_id = aws_api_gateway_method.hack_week_proxy.resource_id
#  http_method = aws_api_gateway_method.hack_week_proxy.http_method
#
#  integration_http_method = "POST"
#  type                    = "AWS_PROXY"
#  uri                     = aws_lambda_function.hack_week_lambda.invoke_arn
#}
#
#resource "aws_api_gateway_method" "proxy_root" {
#  rest_api_id   = aws_api_gateway_rest_api.hack_week_gateway.id
#  resource_id   = aws_api_gateway_rest_api.hack_week_gateway.root_resource_id
#  http_method   = "ANY"
#  authorization = "NONE"
#}
#
#resource "aws_api_gateway_integration" "hack_week_root_integration" {
#  rest_api_id = aws_api_gateway_rest_api.hack_week_gateway.id
#  resource_id = aws_api_gateway_method.proxy_root.resource_id
#  http_method = aws_api_gateway_method.proxy_root.http_method
#
#  integration_http_method = "POST"
#  type                    = "AWS_PROXY"
#  uri                     = aws_lambda_function.hack_week_lambda.invoke_arn
#}
#
#resource "aws_api_gateway_deployment" "example" {
#  depends_on = [
#    aws_api_gateway_integration.hack_week_integration,
#    aws_api_gateway_integration.hack_week_root_integration,
#  ]
#
#  rest_api_id = aws_api_gateway_rest_api.hack_week_gateway.id
#  stage_name  = "hack-week"
#}
#
#resource "aws_lambda_permission" "hack_week_permission" {
#  statement_id  = "AllowAPIGatewayInvoke"
#  action        = "lambda:InvokeFunction"
#  function_name = aws_lambda_function.hack_week_lambda.function_name
#  principal     = "apigateway.amazonaws.com"
#
#  source_arn = "${aws_api_gateway_rest_api.hack_week_gateway.execution_arn}/*/*"
#}
