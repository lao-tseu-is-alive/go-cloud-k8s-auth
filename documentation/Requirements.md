## Requirements

The application should be usable as a _web app_ and use the
[twelve-factor app](https://12factor.net/) methodology for building software-as-a-service apps that:

+ Use declarative formats for setup automation, to minimize time and cost for new developers joining the project
+ Have a clean contract with the underlying operating system, offering maximum portability between execution
  environments
+ Are suitable for deployment on modern cloud platforms, obviating the need for servers and systems administration;
+ Minimize divergence between development and production, enabling continuous deployment for maximum agility;
+ Allow to scale up without significant changes to tooling, architecture, or development practices.

### Functional requirements (FR)

1. Allow CRUD operation on goCloudAuth.
2. List goCloudAuths filtering by type and creator.
3. Search goCloudAuths by keywords, type, creator.
4. Allow retrieving a goCloudAuth by an external Id.
5. Restrict goCloudAuth Create,Modify and Delete to some groups/roles by Type of goCloudAuths, and for specific goCloudAuth.
6. Get number (count) of goCloudAuths by type.
7. Allow CRUD operation on TypegoCloudAuth.
8. List TypegoCloudAuth filtering by creator.
9. Restrict TypegoCloudAuth Create,Modify and Delete to some the role goCloudAuth Administrator.
10. Get number (count) of TypegoCloudAuths by type.
11. Keep automatically track of when and who creates a goCloudAuth.
12. Keep automatically track of when was done the last modification to a goCloudAuth and who has done it
13. Keep automatically track of when and who  deletes a goCloudAuth, mark the goCloudAuth as deleted but do not remove the record from the database. The usual list and search should not return a record of goCloudAuth marked for deletion
15. Include the current status of a goCloudAuth as an Enum attribute
16. Allow a record of a goCloudAuth to be "inactive"
17. Keep track of when and who someone mark a goCloudAuth as inactive, allow to enter an inactivation reason
18. Allow a record of a goCloudAuth to be "validated"
19. Keep track of when and who someone validates a goCloudAuth
20. Allow a goCloudAuth to have an attribute managed_by
21. Keep automatically track of when and who creates a TypegoCloudAuth.
22. Keep automatically track of when was done the last modification to a TypegoCloudAuth and who has done it
23. Keep automatically track of when and who  deletes a TypegoCloudAuth, mark the goCloudAuth as deleted but do not remove the record from the database. The usual list and search should not return a record of TypegoCloudAuth marked for deletion.
24. Allow a record of a TypegoCloudAuth to be "inactive"
25. Keep track of when and who someone mark a TypegoCloudAuth as inactive, allow to enter an inactivation reason

### System requirements (SR)

To ensure the success, reliability, and security of this modern web application we have this constraints : .

1. **Performance:**
    - **Response Time:** The application should provide fast response times to user actions, ensuring a smooth and
      efficient user experience.
    - **Throughput:** The system should handle a certain number of concurrent users or transactions without performance
      degradation.
    - **Scalability:** The ability to scale up or out to accommodate increased load as the user base grows (see §5).
    - **Load Handling:** The application should handle peak loads without crashing or slowing down significantly.

2. **Reliability:**
    - **Availability:** The application should be available and accessible to users as per the agreed-upon uptime
      percentage.
    - **Fault Tolerance:** The application should continue to function or gracefully degrade in the presence of
      failures.
    - **Recovery:** The system should have mechanisms to recover from failures, including data backups and disaster
      recovery plans.

3. **Security:**
    - **Authentication and Authorization:** Only authorized users should have access to specific features and data
      using [JSON Web Tokens : JWT-RFC 7519](https://jwt.io/).
    - **Data Protection:** Sensitive data should be safeguarded through access controls.
    - **Vulnerability Management:** Regular identification, assessment, and mitigation of potential security
      vulnerabilities (via APi Gateway or Service Mesh).

4. **Usability:**
    - **User Interface (UI) and User Experience (UX):** The application's UI should be intuitive, user-friendly, and
      responsive across devices.
    - **Accessibility:** The application should be usable by people with disabilities, conforming to accessibility
      standards.

5. **Scalability:**
    - **Horizontal and Vertical Scaling:** The ability to add more resources (horizontal scaling) and increase resource
      capacity (vertical scaling) as needed.
    - **Elasticity:** The system should automatically scale based on load, optimizing resource utilization.

6. **Maintainability:**
    - **Modularity and Extensibility:** The application's architecture should be modular, allowing new features without
      affecting other parts.
    - **Code Quality:** Well-structured, documented code adhering to standards for easier understanding and maintenance.
    - **Version Control:** Use of version control systems to track changes and facilitate collaboration.

7. **Interoperability:**
    - **APIs and Integration:** The application should seamlessly integrate with other systems through well-defined
      APIs.
    - **Compatibility:** The application should work across browsers, devices, and operating systems.

8. **Performance Efficiency:**
    - **Resource Utilization:** Efficient use of system resources (CPU, memory, network) to avoid bottlenecks and ensure
      optimal performance.
    - **Caching:** Effective use of caching mechanisms to reduce backend load and improve response times.

9. **Compliance:**
    - **Regulatory and Legal Requirements:** The application should comply with industry regulations and legal
      standards.

10. **Operational and Support:**
    - **Monitoring and Logging:** Robust monitoring and logging mechanisms to track application health, performance, and
      issues.
    - **Documentation:** Comprehensive documentation for developers, administrators, and users to aid troubleshooting
      and usage.

