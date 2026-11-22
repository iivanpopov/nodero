use std::{error::Error, fs, process::Command};
use tracing::info;

pub struct Rule {
    pub id: String,
    pub name: String,
    pub pattern: String,
}

impl Rule {
    pub fn to_xml(&self) -> String {
        format!(
            r#"<FilePathRule
              Id="{}"
              Name="{}"
              Description=""
              UserOrGroupSid="S-1-1-0"
              Action="Deny">
              <Conditions>
                  <FilePathCondition Path="{}" />
              </Conditions>
            </FilePathRule>"#,
            &self.id, &self.name, &self.pattern
        )
    }
}

const XML_PATH: &str = "C:\\policy.xml";

pub fn apply_rules(rules: &[&Rule]) -> Result<(), Box<dyn Error>> {
    let rules_string = rules
        .iter()
        .map(|r| r.to_xml())
        .collect::<Vec<String>>()
        .join("\n");

    let policy = format!(
        r#"
        <AppLockerPolicy Version="1">
          <RuleCollection Type="Exe" EnforcementMode="Enabled">
            <FilePathRule
              Id="c918f361-a82f-4c68-b790-113c289bc012"
              Name="Allow all executables"
              Description=""
              UserOrGroupSid="S-1-1-0"
              Action="Allow">
              <Conditions>
                <FilePathCondition Path="*" />
              </Conditions>
            </FilePathRule>
            {}
          </RuleCollection>

          <RuleCollection Type="Script" EnforcementMode="Enabled">
            <FilePathRule
              Id="a98b4f12-6c45-4e21-912a-d45f78e301bc"
              Name="Allow all scripts"
              Description=""
              UserOrGroupSid="S-1-1-0"
              Action="Allow">
              <Conditions>
                <FilePathCondition Path="*" />
              </Conditions>
            </FilePathRule>
          </RuleCollection>

          <RuleCollection Type="Msi" EnforcementMode="Enabled">
            <FilePathRule
              Id="b15d2f47-8e63-4c90-a125-f76e203d91ab"
              Name="Allow all installers"
              Description=""
              UserOrGroupSid="S-1-1-0"
              Action="Allow">
              <Conditions>
                <FilePathCondition Path="*" />
              </Conditions>
            </FilePathRule>
          </RuleCollection>

          <RuleCollection Type="Appx" EnforcementMode="Enabled">
            <FilePublisherRule
              Id="d29c6a81-3f54-4b17-908e-a62d71f4c035"
              Name="Allow all packaged apps"
              Description=""
              UserOrGroupSid="S-1-1-0"
              Action="Allow">
              <Conditions>
                <FilePublisherCondition
                  PublisherName="*"
                  ProductName="*"
                  BinaryName="*">
                  <BinaryVersionRange
                    LowSection="0.0.0.0"
                    HighSection="*" />
                </FilePublisherCondition>
              </Conditions>
            </FilePublisherRule>
          </RuleCollection>
        </AppLockerPolicy>
        "#,
        rules_string
    );

    info!(policy, "Applying policy");

    fs::write(XML_PATH, policy)?;

    Command::new("powershell.exe")
        .args([
            "-NoLogo",
            "-NoProfile",
            "-NonInteractive",
            "-Command",
            &format!("Set-AppLockerPolicy -XmlPolicy '{XML_PATH}' -ErrorAction Stop"),
        ])
        .output()?;

    fs::remove_file(XML_PATH)?;

    Ok(())
}
