package section


func area_4node(na,nb, nc, nd Node ) float64 {
    return 2*area_3node(na,nb,nc);//fabs(0.5*((nb.x - na.x)*(nc.y - na.y)- (nc.x - na.x)*(nb.y - na.y)))*2;
};

type Shape struct {
}

float64 Shape::X_Zero(Node n1, Node n2)
{
    float64 a = (n2.y-n1.y)/(n2.x-n1.x);
    float64 b = n1.y - a*n1.x;
    return -b/a;
};

float64 Shape::Y_Zero(Node n1, Node n2)
{
    float64 a = (n2.y-n1.y)/(n2.x-n1.x);
    float64 b = n1.y - a*n1.x;
    return b;
};

float64 Shape::Jx_node(Node n0, Node n1, Node n2)
{
    Node temp_n[3];
    temp_n[0] = n0;
    temp_n[1] = n1;
    temp_n[2] = n2;
    float64 a = area_3node(n0,n1,n2);
    float64 X_MIN = temp_n[0].x;
    float64 Y_MIN = temp_n[0].y;
    for(int i=0;i<3;i++)
    {
        if(Y_MIN > temp_n[i].y)  Y_MIN = temp_n[i].y;
        if(X_MIN > temp_n[i].x)  X_MIN = temp_n[i].x;
    }

    for(int i=0;i<3;i++)
    {
        temp_n[i].x -= X_MIN;
        temp_n[i].y -= Y_MIN;
    }

    type_LLU x_left = 0, x_mid = 0, x_right =0;
    if(temp_n[0].x >= temp_n[1].x && temp_n[0].x > temp_n[2].x)
    {
        x_right = 0;
        if(temp_n[1].x > temp_n[2].x) { x_mid = 1; x_left = 2;}
        else                          { x_mid = 2; x_left = 1;}
    };
    if(temp_n[1].x >= temp_n[0].x && temp_n[1].x > temp_n[2].x)
    {
        x_right = 1;
        if(temp_n[0].x > temp_n[2].x) { x_mid = 0; x_left = 2;}
        else                          { x_mid = 2; x_left = 0;}
    };
    if(temp_n[2].x >= temp_n[1].x && temp_n[2].x > temp_n[0].x)
    {
        x_right = 2;
        if(temp_n[0].x > temp_n[1].x) { x_mid = 0; x_left = 1;}
        else                          { x_mid = 1; x_left = 0;}
    };
    if(temp_n[x_left].x == temp_n[x_mid].x && temp_n[x_left].y < temp_n[x_mid].y)
    {
        type_LLU r = x_left;
        x_left  = x_mid;
        x_mid   = r;
    }
    if(temp_n[x_right].x == temp_n[x_mid].x && temp_n[x_right].y < temp_n[x_mid].y)
    {
        type_LLU r = x_right;
        x_right = x_mid;
        x_mid   = r;
    }

    type_LLU type  = 0;
    float64 y0 =  temp_n[x_left ].y + (temp_n[x_right].y-temp_n[x_left ].y)/
                (temp_n[x_right].x-temp_n[x_left ].x)*(temp_n[x_mid].x-temp_n[x_left].x);
    if(temp_n[x_mid].y < y0) type = 0;
    else type = 1;

    float64 jx = -1e30;
    float64 Jx_left_mid    = Jx_node(temp_n[x_left ], temp_n[x_mid  ]);
    float64 Jx_mid_right   = Jx_node(temp_n[x_mid  ], temp_n[x_right]);
    float64 Jx_left_right  = Jx_node(temp_n[x_left ], temp_n[x_right]);


    if(type == 0)
    {
        jx  = +Jx_left_right
              -Jx_left_mid
              -Jx_mid_right;
    }

    if(type == 1)
    {
        jx  = -Jx_left_right
              +Jx_left_mid
              +Jx_mid_right;
    }




    float64 YC = (temp_n[0].y+temp_n[1].y+temp_n[2].y)/3.;
    if(jx < 1e-10) jx = 0;
    if(jx <0) {print_name("jx is less NULL");printf("jx[%e]\n",jx);}
    if(a  <0) print_name("area is less NULL");
    jx += -a*pow(YC,2.)+a*pow(Y_MIN+YC,2.);//*fabs(Y_MIN)/Y_MIN;

    return jx;
}

float64 Shape::Jx_node(Node n1, Node n2)
{

    Node temp_n1 = n1;
    Node temp_n2 = n2;
    if(n1.y == 0 && n2.y == 0)
        return 0;
    if(n1.x == n2.x)
        return 0;
    if(n1.x == n2.x && n1.y == n2.y)
    {
        print_name("STRANGE");
        WARNING();
        return 0;
    }

    if(temp_n1.x > temp_n2.x)
    {
        swap(temp_n1.x,temp_n2.x);
        return Jx_node(temp_n1,temp_n2);
    }
    if(temp_n1.y > temp_n2.y)
    {
        swap(temp_n1.y,temp_n2.y);
        return Jx_node(temp_n1,temp_n2);
    }

    float64 jx = 0;
    float64 a = temp_n1.y;
    float64 b = fabs(temp_n2.x - temp_n1.x);
    float64 h = fabs(temp_n2.y - temp_n1.y);
    if(temp_n2.y < a) print_name("WARNING: position a");
    jx = (b*pow(a,3.)/12.+a*b*pow(a/2.,2.))+(b*pow(h,3.)/12.+(b*h/2.)*pow(a,2.));
    return fabs(jx);
};

float64 Shape::CalcJ(float64 Angle)
{
    float64 J = 0;
    mesh->RotatePointXOY(0,0,Angle);
    for(type_LLU i=0;i<mesh->elements.GetSize();i++)
    {
        Element el = mesh->elements.Get(i);
        if(el.ElmType == ELEMENT_TYPE_TRIANGLE)
        {
            Node p[3];
            p[0] = mesh->nodes.Get(el.node[0]-1);
            p[1] = mesh->nodes.Get(el.node[1]-1);
            p[2] = mesh->nodes.Get(el.node[2]-1);
            J += Jx_node(p[0],p[1],p[2]);
        }
    }
    mesh->RotatePointXOY(0,0,-Angle);
    return J;
}

float64 Shape::AngleWithMinimumJ(float64 step0, float64 _angle)
{
    float64 x0 = _angle-step0*1;
    float64 x1 = _angle+step0*0;
    float64 x2 = _angle+step0*1;
    float64 y0 = CalcJ(x0);
    float64 y1 = CalcJ(x1);
    float64 y2 = CalcJ(x2);
    float64 eps= 1e-6;
    if(GRADIANS(max(x0,x1,x2)-min(x0,x1,x2))<=eps || max(y0,y1,y2)-min(y0,y1,y2)<=eps*min(y0,y1,y2))
    {
             if(y0 == min(y0,y1,y2)) return x0;
        else if(y1 == min(y0,y1,y2)) return x1;
        else return x2;
    }
         if(min(y0,y1,y2) == y0) return AngleWithMinimumJ(step0, x0);
    else if(min(y0,y1,y2) == y2) return AngleWithMinimumJ(step0, x2);
    else                         return AngleWithMinimumJ(step0/1.5, x1);
}

void Shape::Calculate()
{
    type_LLU i,numEl = mesh->elements.GetSize();
    float64 Xc = 0;
    float64 Yc = 0;
    Area  = 0;
    for(i=0;i<numEl;i++)
    {
        Element el = mesh->elements.Get(i);
        if(el.ElmType == ELEMENT_TYPE_TRIANGLE)
            {
                Node p[3];
                p[0] = mesh->nodes.Get(el.node[0]-1);
                p[1] = mesh->nodes.Get(el.node[1]-1);
                p[2] = mesh->nodes.Get(el.node[2]-1);
                float64 xc,yc;
                xc = (p[0].x+p[1].x+p[2].x)/3.;
                yc = (p[0].y+p[1].y+p[2].y)/3.;
                float64 a = area_3node(p[0],p[1],p[2]);
                Xc = (a*xc + Area*Xc)/(Area+a);
                Yc = (a*yc + Area*Yc)/(Area+a);
                Area +=a;
            }
    }
    type_LLU numPoint =  mesh->nodes.GetSize();
    for(i=0;i<numPoint;i++)
    {
        mesh->nodes.Get(i).x -= Xc;
        mesh->nodes.Get(i).y -= Yc;
    }

    {
        float64 Ymax  = 0;
        float64 angle = 0;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<mesh->nodes.GetSize();i++)
            Ymax = max(Ymax,fabs(mesh->nodes.Get(i).y));
        Jx_MomentInertia = CalcJ(0);
        Wx_MomentInertia = Jx_MomentInertia/Ymax;
        mesh->RotatePointXOY(0,0,-RADIANS(angle));
    }
    {
        float64 Ymax  = 0;
        float64 angle = 90;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<mesh->nodes.GetSize();i++)
            Ymax = max(Ymax,fabs(mesh->nodes.Get(i).y));
        Jy_MomentInertia = CalcJ(0);
        Wy_MomentInertia = Jy_MomentInertia/Ymax;
        mesh->RotatePointXOY(0,0,-RADIANS(angle));
    }
    float64 AngleMinJ = GRADIANS(AngleWithMinimumJ(RADIANS(45.),0.));
    {
        float64 Ymax  = 0;
        float64 angle = AngleMinJ;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<mesh->nodes.GetSize();i++)
            Ymax = max(Ymax,fabs(mesh->nodes.Get(i).y));
        Jv_MomentInertia = CalcJ(0);
        Wv_MomentInertia = Jv_MomentInertia/Ymax;
        mesh->RotatePointXOY(0,0,-RADIANS(angle));
    }
    {
        float64 Ymax  = 0;
        float64 angle = AngleMinJ+90;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<mesh->nodes.GetSize();i++)
            Ymax = max(Ymax,fabs(mesh->nodes.Get(i).y));
        Ju_MomentInertia = CalcJ(0);
        Wu_MomentInertia = Ju_MomentInertia/Ymax;
        mesh->RotatePointXOY(0,0,-RADIANS(angle));
    }
    // WX_PLASTIC //
    {
        Wx_Plastic = 0;
        float64 angle = 0;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<numEl;i++)
        {
            Element el = mesh->elements.Get(i);
            if(el.ElmType == ELEMENT_TYPE_TRIANGLE)
                {
                    Node p[3];
                    p[0] = mesh->nodes.Get(el.node[0]-1);
                    p[1] = mesh->nodes.Get(el.node[1]-1);
                    p[2] = mesh->nodes.Get(el.node[2]-1);
                    bool simple_case = false;
                    if(p[0].y != 0 || p[1].y != 0 || p[2].y != 0)
                    if(p[0].y/p[1].y >0 && p[0].y/p[2].y >0 && p[1].y/p[2].y > 0)
                    {
                        Wx_Plastic += (fabs(p[0].y+p[1].y+p[2].y)/3.)*(area_3node(p[0],p[1],p[2]));
                        simple_case = true;
                    }
                    if(!simple_case)
                    {
                        if  ((p[0].y == 0 || p[1].y == 0)||
                             (p[1].y == 0 || p[2].y == 0)||
                             (p[2].y == 0 || p[0].y == 0))
                        {
                            Wx_Plastic += (fabs(p[0].y+p[1].y+p[2].y)/3.)*(area_3node(p[0],p[1],p[2]));
                        }
                        else if(p[0].y == 0 && p[1].y/p[2].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            tmp.y = tmp.z = 0.;
                            Wx_Plastic += (fabs(p[0].y+p[1].y+tmp.y)/3.)*(area_3node(p[0],p[1],tmp));
                            Wx_Plastic += (fabs(p[0].y+p[2].y+tmp.y)/3.)*(area_3node(p[0],p[2],tmp));
                        }
                        else if(p[1].y == 0 && p[2].y/p[0].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            tmp.y = tmp.z = 0.;
                            Wx_Plastic += (fabs(p[1].y+p[0].y+tmp.y)/3.)*(area_3node(p[1],p[0],tmp));
                            Wx_Plastic += (fabs(p[1].y+p[2].y+tmp.y)/3.)*(area_3node(p[1],p[2],tmp));
                        }
                        else if(p[2].y == 0 && p[1].y/p[0].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            tmp.y = tmp.z = 0.;
                            Wx_Plastic += (fabs(p[2].y+p[0].y+tmp.y)/3.)*(area_3node(p[2],p[0],tmp));
                            Wx_Plastic += (fabs(p[2].y+p[1].y+tmp.y)/3.)*(area_3node(p[2],p[1],tmp));
                        }
                        else if(p[0].y/p[1].y < 0 && p[0].y/p[2].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            Wx_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[1].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[1],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[2].y+p[1].y+tmp2.y)/3.)*(area_3node(p[2],p[1],tmp2));
                        }
                        else if(p[1].y/p[0].y < 0 && p[1].y/p[2].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            Wx_Plastic += (fabs(p[1].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[1],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[2].y+p[0].y+tmp2.y)/3.)*(area_3node(p[2],p[0],tmp2));
                        }
                        else if(p[2].y/p[0].y < 0 && p[2].y/p[1].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            Wx_Plastic += (fabs(p[2].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[2],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wx_Plastic += (fabs(p[1].y+p[0].y+tmp2.y)/3.)*(area_3node(p[1],p[0],tmp2));
                        }
                        else
                        printf("IO-");
                    }
                }
            }
            mesh->RotatePointXOY(0,0,-RADIANS(angle));
        }
    {
        Wy_Plastic = 0;
        float64 angle = 90;
        mesh->RotatePointXOY(0,0,+RADIANS(angle));
        for(i=0;i<numEl;i++)
        {
            Element el = mesh->elements.Get(i);
            if(el.ElmType == ELEMENT_TYPE_TRIANGLE)
                {
                    Node p[3];
                    p[0] = mesh->nodes.Get(el.node[0]-1);
                    p[1] = mesh->nodes.Get(el.node[1]-1);
                    p[2] = mesh->nodes.Get(el.node[2]-1);
                    bool simple_case = false;
                    if(p[0].y != 0 || p[1].y != 0 || p[2].y != 0)
                    if(p[0].y/p[1].y >0 && p[0].y/p[2].y >0 && p[1].y/p[2].y > 0)
                    {
                        Wy_Plastic += (fabs(p[0].y+p[1].y+p[2].y)/3.)*(area_3node(p[0],p[1],p[2]));
                        simple_case = true;
                    }
                    if(!simple_case)
                    {
                        if  ((p[0].y == 0 || p[1].y == 0)||
                             (p[1].y == 0 || p[2].y == 0)||
                             (p[2].y == 0 || p[0].y == 0))
                        {
                            Wy_Plastic += (fabs(p[0].y+p[1].y+p[2].y)/3.)*(area_3node(p[0],p[1],p[2]));
                        }
                        else if(p[0].y == 0 && p[1].y/p[2].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            tmp.y = tmp.z = 0.;
                            Wy_Plastic += (fabs(p[0].y+p[1].y+tmp.y)/3.)*(area_3node(p[0],p[1],tmp));
                            Wy_Plastic += (fabs(p[0].y+p[2].y+tmp.y)/3.)*(area_3node(p[0],p[2],tmp));
                        }
                        else if(p[1].y == 0 && p[2].y/p[0].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            tmp.y = tmp.z = 0.;
                            Wy_Plastic += (fabs(p[1].y+p[0].y+tmp.y)/3.)*(area_3node(p[1],p[0],tmp));
                            Wy_Plastic += (fabs(p[1].y+p[2].y+tmp.y)/3.)*(area_3node(p[1],p[2],tmp));
                        }
                        else if(p[2].y == 0 && p[1].y/p[0].y < 0)
                        {
                            Node tmp; // intersect with zero line
                            tmp.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            tmp.y = tmp.z = 0.;
                            Wy_Plastic += (fabs(p[2].y+p[0].y+tmp.y)/3.)*(area_3node(p[2],p[0],tmp));
                            Wy_Plastic += (fabs(p[2].y+p[1].y+tmp.y)/3.)*(area_3node(p[2],p[1],tmp));
                        }
                        else if(p[0].y/p[1].y < 0 && p[0].y/p[2].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            Wy_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[1].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[1],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[2].y+p[1].y+tmp2.y)/3.)*(area_3node(p[2],p[1],tmp2));
                        }
                        else if(p[1].y/p[0].y < 0 && p[1].y/p[2].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[1].x-p[0].x)*fabs(p[0].y)/(fabs(p[1].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            Wy_Plastic += (fabs(p[1].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[1],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[2].y+p[0].y+tmp2.y)/3.)*(area_3node(p[2],p[0],tmp2));
                        }
                        else if(p[2].y/p[0].y < 0 && p[2].y/p[1].y < 0 )
                        {
                            Node tmp1; tmp1.y = tmp1.z = 0.;
                            tmp1.x = p[0].x+(p[2].x-p[0].x)*fabs(p[0].y)/(fabs(p[2].y)+fabs(p[0].y));
                            Node tmp2; tmp2.y = tmp2.z = 0.;
                            tmp2.x = p[1].x+(p[2].x-p[1].x)*fabs(p[1].y)/(fabs(p[2].y)+fabs(p[1].y));
                            Wy_Plastic += (fabs(p[2].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[2],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[0].y+tmp1.y+tmp2.y)/3.)*(area_3node(p[0],tmp1,tmp2));
                            Wy_Plastic += (fabs(p[1].y+p[0].y+tmp2.y)/3.)*(area_3node(p[1],p[0],tmp2));
                        }
                        else
                        printf("IO-");
                    }
                }
        }
        mesh->RotatePointXOY(0,0,-RADIANS(angle));
    }
}
