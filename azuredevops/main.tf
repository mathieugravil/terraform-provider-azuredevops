resource "azuredevops_project" "test" {
  name              = "testgo"
  description       = " md "
  visibility        = "Private"
  sourceControlType = "Git"
  templateTypeId    = "adcc42ab-9882-485e-a3ed-7678f01f66bc"
}