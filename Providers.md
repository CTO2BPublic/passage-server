# AWS

## 1. How to Get `InstanceARN` and `IdentityStoreID` for AWS Identity Center

### **IdentityStoreID**
The **Identity Store ID** is unique to your AWS organization and is tied to AWS Identity Center.

#### Steps:
1. **Log in to the AWS Management Console.**
2. Navigate to **AWS Identity Center** (search for "Identity Center" in the Services section).
3. In the left-hand menu, go to **Settings**.
4. Scroll down to the **Identity Source** section:
   - If using AWS-managed users, the **Identity Store ID** will be displayed here.
   - If using an external identity source (e.g., Active Directory or an external IdP), the ID will also appear in this section.

---

### **InstanceARN**
The **Instance ARN** refers to the ARN of the AWS Identity Center instance.

#### Steps:
1. Open your terminal and execute the following AWS CLI command:
   ```bash
   aws sso-admin list-instances
