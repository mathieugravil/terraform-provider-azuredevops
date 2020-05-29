resource "azuredevops_project" "test" {
  name              = "testgo"
  description       = " md "
  visibility        = "Private"
  source_control_type = "Git"
  template_type_id    = "adcc42ab-9882-485e-a3ed-7678f01f66bc"
}