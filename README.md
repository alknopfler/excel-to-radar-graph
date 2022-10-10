# Excel To Radar graph

Based on an excel existing you could use this repo to upload, process, and generate the radar graph with this information:

- config/example.yaml 
```
Data:
  - Name: "First Category"
    Total: 35    //The total range of points possible
    List:        //Sum of cells to reach the Total
      - B2       //Excel Cells with value to be count in order to reach the total
      - D2
      - E2
      - F2
      - H2
      - I2
      - K2       //The sum of all of this cells will be the current value against the total value.

  - Name: "Category 2"
    Total: 10
    List:
      - M2
      - O2
```
- website to upload an existing excel 
Pointing to http://localhost a basic website will be opened just to upload the excel file in order to be processed.

- chart graphics
As a result of the process described before, the graphic will be generated just to copy/download from the website pointing to localhost.

